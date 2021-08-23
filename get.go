package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
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

func getCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 1 || argsLength > 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
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
		return cli.NewExitError(fmt.Sprintf("Failed to get data. \n%v", err), 82)
	}

	writeSnapshot(c.App.Writer, docsnap, extendedJson)

	return nil
}
