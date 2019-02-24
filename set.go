package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func setData(
	client *firestore.Client,
	collection string,
	id string,
	data string) error {

	object, err := unmarshallData(data)
	if err != nil {
		return err
	}

	_, err = client.
		Collection(collection).
		Doc(id).
		Set(context.Background(), object)

	if err != nil {
		return err
	}

	return nil
}

func setCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()
	id := c.Args().Get(1)
	data := c.Args().Get(2)
	client, err := createClient(credentials)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
	}
	err = setData(client, collectionPath, id, data)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to write data. \n%v", err), 80)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", id)
	defer client.Close()
	return nil
}
