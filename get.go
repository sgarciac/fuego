package main

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"github.com/urfave/cli/v3"
)

func getData(
	client *firestore.Client,
	collectionPath string,
	documentPath string,
	id string,
) (*firestore.DocumentSnapshot, error) {

	var documentRef *firestore.DocumentRef
	if collectionPath != "" {
		collectionRef := client.Collection(collectionPath)
		documentRef = collectionRef.Doc(id)
	} else {
		documentRef = client.Doc(documentPath)
	}
	return documentRef.Get(context.Background())
}

func getCommandAction(ctx context.Context, c *cli.Command) error {
	argsLength := c.Args().Len()

	if argsLength < 1 || argsLength > 2 {
		return cli.Exit("Wrong number of arguments", 82)
	}

	extendedJson := c.Bool("extendedjson")

	var collectionPath, documentPath, id string

	if argsLength == 2 {
		collectionPath = c.Args().First()
		id = c.Args().Get(1)
	} else {
		documentPath = c.Args().First()
	}

	client, err := createClient(credentials)

	if err != nil {
		return cliClientError(err)
	}

	defer client.Close()

	docsnap, err := getData(client, collectionPath, documentPath, id)

	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get data. \n%v", err), 82)
	}

	writeSnapshot(c.Root().Writer, docsnap, extendedJson)

	return nil
}
