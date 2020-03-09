package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func setData(
	client *firestore.Client,
	collection string,
	id string,
	data string,
	timestampify bool,
	merge bool) error {

	object, err := unmarshallData(data)
	if err != nil {
		return err
	}

	if timestampify {
		timestampifyMap(object)
	}

	var options []firestore.SetOption
	if merge {
		options = append(options, firestore.MergeAll)
	}

	_, err = client.
		Collection(collection).
		Doc(id).
		Set(context.Background(), object, options...)

	if err != nil {
		return err
	}

	return nil
}

func setCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()
	timestampify := c.Bool("timestamp")
	merge := c.Bool("merge")

	id := c.Args().Get(1)
	data := c.Args().Get(2)
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	err = setData(client, collectionPath, id, data, timestampify, merge)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to write data. \n%v", err), 85)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", id)
	defer client.Close()
	return nil
}
