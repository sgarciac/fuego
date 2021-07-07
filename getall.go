package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"strings"
)

func getDocuments(client *firestore.Client, collectionPath string, ids []string) (string, error) {
	var data []map[string]interface{}

	if collectionPath != "" {
		collectionRef := client.Collection(collectionPath)

		var docRefs []*firestore.DocumentRef
		for _, elem := range ids {
			docRefs = append(docRefs, collectionRef.Doc(elem))
		}

		all, err := client.GetAll(context.Background(), docRefs)
		if err != nil {
			return "", err
		}
		for _, docSnap := range all {
			elem := docSnap.Data()
			if elem != nil {
				data = append(data, elem)
			}
		}
	}
	result, err := json.MarshalIndent(data, "", "  ")
	return string(result), err
}

func getAllCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 1 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	var collectionPath string
	var id []string
	collectionPath = c.Args().First()
	id = strings.Split(c.Args()[1], ",")
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	data, err := getDocuments(client, collectionPath, id)

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get data. \n%v", err), 82)
	}

	fmt.Fprintf(c.App.Writer, "%v\n", data)

	defer client.Close()
	return nil
}
