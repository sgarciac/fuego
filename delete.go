package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
)

func deleteCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()
	id := c.Args().Get(1)
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	collectionRef := client.Collection(collectionPath)
	documentRef := collectionRef.Doc(id)
	res, err := documentRef.Delete(context.Background())
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to delete data. \n%v", res), 82)
	}
	defer client.Close()
	return nil
}
