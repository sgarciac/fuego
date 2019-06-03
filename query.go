package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"gopkg.in/urfave/cli.v1"
	"io"
	"log"
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

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()

	// pagination
	startAt := c.String("startat")
	startAfter := c.String("startafter")
	endAt := c.String("endat")
	endBefore := c.String("endbefore")
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

	collectionRef := client.Collection(collectionPath)
	if limit < batch {
		batch = limit
	}
	query := collectionRef.Limit(batch)

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
		documentRef := collectionRef.Doc(startAt)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", startAt), 83)
		}
		query = query.StartAt(docsnap)
	}

	if startAfter != "" {
		documentRef := collectionRef.Doc(startAfter)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", startAfter), 83)
		}
		query = query.StartAfter(docsnap)
	}

	if endAt != "" {
		documentRef := collectionRef.Doc(endAt)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", endAt), 83)
		}
		query = query.EndAt(docsnap)
	}

	if endBefore != "" {
		documentRef := collectionRef.Doc(endBefore)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", endBefore), 83)
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
			retrieved ++

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
