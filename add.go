package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func addData(
	client *firestore.Client,
	collection string,
	data string) (string, error) {

	object, err := unmarshallData(data)
	if err != nil {
		return "", err
	}

	doc, _, err := client.
		Collection(collection).
		Add(context.Background(), object)

	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func addCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()
	client, err := createClient(credentials)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
	}
	id, err := addData(client, collectionPath, c.Args().Get(1))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to add data. \n%v", err), 80)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", id)
	defer client.Close()
	return nil
}
