package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"strings"
)

func deleteAllCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	var collectionPath string
	var ids []string
	collectionPath = c.Args().First()
	ids = strings.Split(c.Args()[1], ",")
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	_, err = deleteAllDocuments(client, collectionPath, ids)

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to remove all documents. \n%v", err), 82)
	}

	defer client.Close()
	return nil
}

func deleteAllDocuments(client *firestore.Client, collectionPath string, ids []string) ([]*firestore.WriteResult, error) {
	batch := client.Batch()
	collectionRef := client.Collection(collectionPath)
	for _, id := range ids {
		batch.Delete(collectionRef.Doc(id))
	}
	return batch.Commit(context.Background())
}
