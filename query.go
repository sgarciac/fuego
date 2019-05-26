package main

import (
	firestore "cloud.google.com/go/firestore"
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
	orderbyFields := c.StringSlice("orderby")
	orderdirFields := c.StringSlice("orderdir")

	queryParser := getQueryParser()
	fieldPathParser := getFieldPathParser()

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	collectionRef := client.Collection(collectionPath)
	query := collectionRef.Limit(c.Int("limit"))

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
		if i < len(orderbyFields) {
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

	documentIterator := query.Documents(context.Background())

	docs, err := documentIterator.GetAll()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
	}

	var displayItems []map[string]interface{}
	for _, doc := range docs {
		var displayItem = make(map[string]interface{})
		displayItem["ID"] = doc.Ref.ID
		displayItem["CreateTime"] = doc.CreateTime
		displayItem["ReadTime"] = doc.ReadTime
		displayItem["UpdateTime"] = doc.UpdateTime
		displayItem["Data"] = doc.Data()
		displayItems = append(displayItems, displayItem)
	}

	jsonString, _ := marshallData(displayItems)
	fmt.Fprintln(c.App.Writer, jsonString)
	return nil
}
