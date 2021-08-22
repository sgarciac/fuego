package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func getDocuments(client *firestore.Client,
	collectionPath string,
	ids []string,
	extendedJson bool,
) ([]*firestore.DocumentSnapshot, error) {
	collectionRef := client.Collection(collectionPath)

	var docRefs []*firestore.DocumentRef
	for _, elem := range ids {
		docRefs = append(docRefs, collectionRef.Doc(elem))
	}

	return client.GetAll(context.Background(), docRefs)
}

func getAllCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	extendedJson := c.Bool("extendedjson")

	var collectionPath string
	var ids []string

	collectionPath = c.Args().First()
	ids = c.Args()[1:]

	client, err := createClient(credentials)

	if err != nil {
		return cliClientError(err)
	}

	data, err := getDocuments(client, collectionPath, ids, extendedJson)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Error fetching documents. \n%v", err), 86)
	}

	displayItemWriter := newDisplayItemWriter(&c.App.Writer)
	defer displayItemWriter.Close()

	for _, doc := range data {
		err = displayItemWriter.Write(doc, extendedJson)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Error while writing output. \n%v", err), 86)
		}
	}

	defer client.Close()
	return nil
}
