package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"google.golang.org/api/iterator"
)

func collectionsCommandAction(ctx context.Context, c *cli.Command) error {
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
			return cli.Exit(fmt.Sprintf("Failed to list collections. \n%v", err), 86)
		}
		fmt.Println(col.ID)
	}
	defer client.Close()
	return nil
}
