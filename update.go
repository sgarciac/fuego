package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func updateData(
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
		Update(context.Background(),
			flattenForUpdate(object, ""))

	if err != nil {
		return err
	}

	return nil
}

func updateCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()

	id := c.Args().Get(1)
	data := c.Args().Get(2)
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	err = updateData(client, collectionPath, id, data)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to update data. \n%v", err), 85)
	}
	fmt.Fprintf(c.App.Writer, "%v\n", id)
	defer client.Close()
	return nil
}

func flattenForUpdate(data map[string]interface{}, root string) (result []firestore.Update) {
	for k, v := range data {
		switch v.(type) {
		case map[string]interface{}:
			result = append(result, flattenForUpdate(v.(map[string]interface{}), k+".")...)
		default:
			result = append(result, firestore.Update{
				Path:      root + k,
				FieldPath: nil,
				Value:     v,
			})
		}
	}

	return
}
