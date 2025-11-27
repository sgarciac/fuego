package main

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"github.com/urfave/cli/v3"
)

func addData(
	client *firestore.Client,
	collection string,
	data string,
) (string, error) {

	object, err := unmarshallData(data)
	if err != nil {
		return "", err
	}

	transformExtendedJsonMapToFirestoreMap(object, client)

	doc, _, err := client.
		Collection(collection).
		Add(context.Background(), object)

	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func addCommandAction(ctx context.Context, c *cli.Command) error {
	collectionPath := c.Args().First()
	data := c.Args().Get(1)

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	id, err := addData(client, collectionPath, data)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to add data. \n%v", err), 81)
	}
	fmt.Fprintf(c.Root().Writer, "%v\n", id)
	defer client.Close()
	return nil
}
