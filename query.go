package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

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
	limit := c.Int("limit")
	batch := c.Int("batch")
	orderby := c.String("orderby")

	queryParser := getQueryParser()
	var queries []Firestorequery = nil
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
		queries = append(queries, parsedQuery)
		query = query.WherePath(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
	}

	// order by
	if orderby != "" {
		var parsedOrderBy Firestorefieldpath
		if err := fieldPathParser.ParseString(orderby, &parsedOrderBy); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing orderby '%s' %v",
				orderby, err), 83)
		}
		query = query.OrderByPath(parsedOrderBy.Key, getDir(c.String("orderdir")))
	} else if queries != nil {
		// if we have a not equality filter we need to use that as ordering
		if queries[0].Operator != "==" {
			query = query.OrderByPath(queries[0].Key, firestore.Asc)
		} else {
			// if we have a equality query we still may have more than `limit` results
			// therefore we set the ordering explicitly to the documentid. without any ordering we
			// would be unable to use later startAt
			query = query.OrderBy(firestore.DocumentID, firestore.Asc)
		}
	} else {
		// default ordering for batched queries must be the documentID
		query = query.OrderBy(firestore.DocumentID, firestore.Asc)
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

	var displayItems []map[string]interface{}
	to_query := limit

	// make queries with a maximum of `batch` results until we have `limit` results or no more documents are returned
	for to_query > 0 {
		documentIterator := query.Documents(context.Background())

		docs, err := documentIterator.GetAll()
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
		}

		var last *firestore.DocumentSnapshot
		for _, doc := range docs {

			var displayItem = make(map[string]interface{})
			last = doc

			displayItem["ID"] = doc.Ref.ID
			displayItem["CreateTime"] = doc.CreateTime
			displayItem["ReadTime"] = doc.ReadTime
			displayItem["UpdateTime"] = doc.UpdateTime
			displayItem["Data"] = doc.Data()
			displayItems = append(displayItems, displayItem)
		}

		if len(docs) == 0 {
			// no more results
			to_query = 0
		} else {
			to_query -= len(docs)
			// if we do not need a complete batch we must adjust the limit
			if to_query < batch {
				query = query.Limit(to_query)
			}
			// we need to figure out what the ordering is and add the correct fields
			if orderby != "" {
				query = query.StartAfter(last.Data()[orderby])
			} else {
				query = query.StartAfter(last.Ref.ID)
			}
		}
	}

	jsonString, _ := marshallData(displayItems)
	_, err = fmt.Fprintln(c.App.Writer, jsonString)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to print result \n%v", err), 85)
	}

	return nil
}
