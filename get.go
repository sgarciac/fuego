package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func getData(
	client *firestore.Client,
	collection string,
	id string) (string, error) {

	collectionRef := client.Collection(collection)
	documentRef := collectionRef.Doc(id)
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
	collectionPath := c.Args().First()
	id := c.Args().Get(1)
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	data, err := getData(client, collectionPath, id)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get data. \n%v", err), 82)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", data)
	defer client.Close()
	return nil
}
