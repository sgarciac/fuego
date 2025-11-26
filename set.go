package main

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	"github.com/urfave/cli/v3"
)

func setData(
	client *firestore.Client,
	collectionPath string,
	documentPath string,
	id string,
	data string,
	merge bool) error {

	object, err := unmarshallData(data)
	if err != nil {
		return err
	}

	transformExtendedJsonMapToFirestoreMap(object, client)

	var options []firestore.SetOption
	if merge {
		options = append(options, firestore.MergeAll)
	}

	if collectionPath != "" {
		_, err = client.
			Collection(collectionPath).
			Doc(id).
			Set(context.Background(), object, options...)
	} else {
		_, err = client.
			Doc(documentPath).
			Set(context.Background(), object, options...)
	}

	if err != nil {
		return err
	}

	return nil
}

func setCommandAction(ctx context.Context, c *cli.Command) error {
	argsLength := c.Args().Len()

	if argsLength < 2 || argsLength > 3 {
		return cli.Exit("Wrong number of arguments", 85)
	}

	merge := c.Bool("merge")

	var collectionPath, id, data, documentPath string

	if argsLength == 3 {
		collectionPath = c.Args().First()
		id = c.Args().Get(1)
		data = c.Args().Get(2)
	} else {
		documentPath = c.Args().First()
		data = c.Args().Get(1)
	}

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	err = setData(client, collectionPath, documentPath, id, data, merge)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write data. \n%v", err), 85)
	}

	if collectionPath != "" {
		fmt.Fprintf(c.Root().Writer, "%v\n", id)
	} else {
		fmt.Fprintf(c.Root().Writer, "%v\n", documentPath)
	}
	defer client.Close()
	return nil
}
