package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func addData(
	client *firestore.Client,
	collection string,
	data string,
	timestampify bool) (string, error) {

	object, err := unmarshallData(data)
	if err != nil {
		return "", err
	}

	if timestampify {
		timestampifyMap(object)
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
	timestampify := c.Bool("timestamp")
	data := c.Args().Get(1)

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	id, err := addData(client, collectionPath, data, timestampify)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to add data. \n%v", err), 81)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", id)
	defer client.Close()
	return nil
}
