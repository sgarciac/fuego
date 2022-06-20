package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"strings"
)

func getDir(name string) firestore.Direction {
	if name == "DESC" {
		return firestore.Desc
	}
	return firestore.Asc
}

func operatorTokenToFirestore(operator string) string {
	switch operator {
	case "<in>":
		return "in"
	case "<not-in>":
		return "not-in"
	case "<array-contains-any>":
		return "array-contains-any"
	case "<array-contains>":
		return "array-contains"
	default:
		return operator
	}
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

	// display
	extendedJson := c.Bool("extendedjson")

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

	queryParser := getQueryParser()

	fieldPathParser := getFieldPathParser()

	client, err := createClient(credentials)

	if err != nil {
		return cliClientError(err)
	}

	// init the query either from a collection ref or a collection group.
	var query firestore.Query
	var collectionRef *firestore.CollectionRef
	var collectionGroupRef *firestore.CollectionGroupRef

	if queryGroup {
		collectionGroupRef = client.CollectionGroup(collectionPathOrId)
		query = collectionGroupRef.Limit(limit)
	} else {
		collectionRef = client.Collection(collectionPathOrId)
		query = collectionRef.Limit(limit)
	}

	// add the conditions one by one.
	for i := 1; i < c.NArg(); i++ {
		queryString := c.Args().Get(i)
		var parsedQuery Firestorequery
		if err := queryParser.ParseString(queryString, &parsedQuery); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing query '%s' %v", queryString, err), 83)
		}
		query = query.WherePath(parsedQuery.Key, operatorTokenToFirestore(parsedQuery.Operator), parsedQuery.Value.get())
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

	// selected fields.
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

	displayItemWriter := newDisplayItemWriter(&c.App.Writer)
	defer displayItemWriter.Close()

	documentIterator := query.Documents(context.Background())

	for {
		doc, err := documentIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			documentIterator.Stop()
			return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
		}

		err = displayItemWriter.Write(doc, extendedJson)
		if err != nil {
			documentIterator.Stop()
			return cli.NewExitError(fmt.Sprintf("Error while writing output. \n%v", err), 86)
		}
	}

	documentIterator.Stop()

	return nil
}
