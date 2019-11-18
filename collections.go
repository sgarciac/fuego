package main

import (
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"github.com/urfave/cli"
)

func collectionsCommandAction(c *cli.Context) error {
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	ci := client.Collections(context.Background())

	for {
		col, err := ci.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to list collections. \n%v", err), 86)
		}
		fmt.Println(col.ID)
	}
	defer client.Close()
	return nil
}
