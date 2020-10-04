package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func deleteCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 1 || argsLength > 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	var documentRef *firestore.DocumentRef

	if argsLength == 2 {
		collectionPath := c.Args().First()
		id := c.Args().Get(1)
		collectionRef := client.Collection(collectionPath)
		documentRef = collectionRef.Doc(id)
	} else {
		documentPath := c.Args().First()
		documentRef = client.Doc(documentPath)
	}

	res, err := documentRef.Delete(context.Background())
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to delete data. \n%v", res), 82)
	}
	defer client.Close()
	return nil
}
