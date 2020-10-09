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
	id string) (string, error) {

	var documentRef *firestore.DocumentRef
	if collectionPath != "" {
		collectionRef := client.Collection(collectionPath)
		documentRef = collectionRef.Doc(id)
	} else {
		documentRef = client.Doc(documentPath)
	}
	docsnap, err := documentRef.Get(context.Background())
	if err != nil {
		return "", err
	}
	jsonString, err := marshallData(docsnap.Data())
	if err != nil {
		return "", err
	}
	return jsonString, nil
}

func getCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 1 || argsLength > 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

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

	data, err := getData(client, collectionPath, documentPath, id)

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get data. \n%v", err), 82)
	}

	fmt.Fprintf(c.App.Writer, "%v\n", data)

	defer client.Close()
	return nil
}
