package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"io"
	"log"
	"strings"
)

// wrapper to stream the json serialized results
type displayItemWriter struct {
	isFirst bool
	writer  *io.Writer
}

func newDisplayItemWriter(writer *io.Writer) displayItemWriter {
	return displayItemWriter{true, writer}
}

func (d *displayItemWriter) Write(doc *firestore.DocumentSnapshot) error {
	if d.isFirst {
		_, err := fmt.Fprintln(*d.writer, "[")
		if err != nil {
			return err
		}
		d.isFirst = false
	} else {
		_, err := fmt.Fprintln(*d.writer, ",")
		if err != nil {
			return err
		}
	}

	var displayItem = make(map[string]interface{})

	displayItem["ID"] = doc.Ref.ID
	displayItem["CreateTime"] = doc.CreateTime
	displayItem["ReadTime"] = doc.ReadTime
	displayItem["UpdateTime"] = doc.UpdateTime
	displayItem["Data"] = doc.Data()

	jsonString, err := marshallData(displayItem)

	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(*d.writer, jsonString)
	if err != nil {
		return err
	}
	return nil
}

func (d *displayItemWriter) Close() {
	if !d.isFirst {
		_, err := fmt.Fprintln(*d.writer, "]")
		if err != nil {
			log.Panicf("Could not write finishing part of results. %v", err)
		}
	}
}

func getDir(name string) firestore.Direction {
	if name == "DESC" {
		return firestore.Desc
	}
	return firestore.Asc
}

// get the snapshot of a document, receiving either a document path, or a
// collection reference and a document id.  The document-path of document id are
// passed in the 'document' parameter.
func documentSnapshot(client *firestore.Client, document string, collectionRef *firestore.CollectionRef, group bool) (*firestore.DocumentSnapshot, error) {
	var documentRef *firestore.DocumentRef

	if group && !strings.Contains(document, "/") {
		return nil, cli.NewExitError("If you use the group option, you must use a document-path for pagination arguments", 83)
	}

	if strings.Contains(document, "/") {
		documentRef = client.Doc(document)
	} else {
		documentRef = collectionRef.Doc(document)
	}
	return documentRef.Get(context.Background())
}

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	collectionPathOrId := c.Args().First()

	// pagination
	startAt := c.String("startat")
	startAfter := c.String("startafter")
	endAt := c.String("endat")
	endBefore := c.String("endbefore")

	queryGroup := c.Bool("group")

	selectFields := c.StringSlice("select")
	orderbyFields := c.StringSlice("orderby")
	orderdirFields := c.StringSlice("orderdir")
	limit := c.Int("limit")
	batch := c.Int("batch")

	queryParser := getQueryParser()
	fieldPathParser := getFieldPathParser()

	client, err := createClient(credentials)

	if err != nil {
		return cliClientError(err)
	}

	if limit < batch {
		batch = limit
	}

	// init the query either from a collection ref or a collection group.
	var query firestore.Query
	var collectionRef *firestore.CollectionRef
	var collectionGroupRef *firestore.CollectionGroupRef

	if queryGroup {
		collectionGroupRef = client.CollectionGroup(collectionPathOrId)
		query = collectionGroupRef.Limit(batch)
	} else {
		collectionRef = client.Collection(collectionPathOrId)
		query = collectionRef.Limit(batch)
	}

	for i := 1; i < c.NArg(); i++ {
		queryString := c.Args().Get(i)
		var parsedQuery Firestorequery
		if err := queryParser.ParseString(queryString, &parsedQuery); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing query '%s' %v", queryString, err), 83)
		}
		query = query.WherePath(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
	}

	// order by
	for i, orderbyRaw := range orderbyFields {
		var parsedOrderBy Firestorefieldpath
		var orderDir string
		if err := fieldPathParser.ParseString(orderbyRaw, &parsedOrderBy); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing orderby '%s' %v",
				orderbyRaw, err), 83)
		}
		if i < len(orderdirFields) {
			orderDir = orderdirFields[i]
		} else {
			orderDir = "DESC"
		}
		query = query.OrderByPath(parsedOrderBy.Key, getDir(orderDir))
	}

	if startAt != "" {
		docsnap, err := documentSnapshot(client, startAt, collectionRef, queryGroup)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s'", startAt), 83)
		}
		query = query.StartAt(docsnap)
	}

	if startAfter != "" {
		docsnap, err := documentSnapshot(client, startAfter, collectionRef, queryGroup)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s'", startAfter), 83)
		}
		query = query.StartAfter(docsnap)
	}

	if endAt != "" {
		docsnap, err := documentSnapshot(client, endAt, collectionRef, queryGroup)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s'", endAt), 83)
		}
		query = query.EndAt(docsnap)
	}

	if endBefore != "" {
		docsnap, err := documentSnapshot(client, endBefore, collectionRef, queryGroup)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s'", endBefore), 83)
		}
		query = query.EndBefore(docsnap)
	}

	if len(selectFields) > 0 {
		var selectFieldPaths []firestore.FieldPath
		for _, selectField := range selectFields {
			var parsedSelect Firestorefieldpath
			if err := fieldPathParser.ParseString(selectField, &parsedSelect); err != nil {
				return cli.NewExitError(fmt.Sprintf("Error parsing select '%s' %v",
					selectField, err), 83)
			}
			selectFieldPaths = append(selectFieldPaths, parsedSelect.Key)
		}
		query = query.SelectPaths(selectFieldPaths...)
	}

	// max amount of documents still to retrieve
	toQuery := limit

	displayItemWriter := newDisplayItemWriter(&c.App.Writer)
	defer displayItemWriter.Close()

	// make queries with a maximum of `batch` results until we have `limit` results or no more documents are returned
	for toQuery > 0 {
		// TODO the Close call should not be defered out of a for loop. Maybe move the iteration into a function and
		// defer there?
		documentIterator := query.Documents(context.Background())
		var last *firestore.DocumentSnapshot
		retrieved := 0

		for {
			doc, err := documentIterator.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				documentIterator.Stop()
				return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
			}

			last = doc
			retrieved++

			err = displayItemWriter.Write(doc)
			if err != nil {
				documentIterator.Stop()
				return cli.NewExitError(fmt.Sprintf("Error while writing output. \n%v", err), 86)
			}
		}

		documentIterator.Stop()

		if retrieved == 0 {
			// no more results
			toQuery = 0
		} else {
			toQuery -= retrieved
			// if we do not need a complete batch we must adjust the max_query
			if toQuery < batch {
				query = query.Limit(toQuery)
			}
			// we need to figure out what the ordering is and add the correct fields
			query = query.StartAfter(last)
		}
	}

	return nil
}
