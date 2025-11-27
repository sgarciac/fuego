package main

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"github.com/urfave/cli/v3"
)

func getDocuments(client *firestore.Client,
	collectionPath string,
	ids []string,

) ([]*firestore.DocumentSnapshot, error) {
	collectionRef := client.Collection(collectionPath)

	var docRefs []*firestore.DocumentRef
	for _, elem := range ids {
		docRefs = append(docRefs, collectionRef.Doc(elem))
	}

	return client.GetAll(context.Background(), docRefs)
}

func getAllCommandAction(ctx context.Context, c *cli.Command) error {
	argsLength := c.Args().Len()

	if argsLength < 2 {
		return cli.Exit("Wrong number of arguments", 82)
	}

	extendedJson := c.Bool("extendedjson")

	var collectionPath string
	var ids []string

	collectionPath = c.Args().First()
	ids = c.Args().Slice()[1:]

	client, err := createClient(credentials)

	if err != nil {
		return cliClientError(err)
	}

	data, err := getDocuments(client, collectionPath, ids)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Error fetching documents. \n%v", err), 86)
	}

	displayItemWriter := newDisplayItemWriter(&c.Root().Writer)
	defer displayItemWriter.Close()

	for _, doc := range data {
		err = displayItemWriter.Write(doc, extendedJson)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Error while writing output. \n%v", err), 86)
		}
	}

	defer client.Close()
	return nil
}
