package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"strings"
)

func deleteAllCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength < 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	var collectionPath string
	var ids []string
	collectionPath = c.Args().First()
	ids = strings.Split(c.Args()[1], ",")
	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}
	deletedCount := 0

	for _, part := range partition(ids, maxWritesCount) {
		d, err := deleteAllDocuments(client, collectionPath, part)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to remove all documents. \n%v", err), 82)
		}
		deletedCount += len(d)
		fmt.Printf("Deleted  %v out of %v\n", deletedCount, len(ids))
	}

	defer client.Close()
	return nil
}

func deleteAllDocuments(client *firestore.Client, collectionPath string, ids []string) ([]*firestore.WriteResult, error) {
	batch := client.Batch()
	collectionRef := client.Collection(collectionPath)
	for _, id := range ids {
		batch.Delete(collectionRef.Doc(id))
	}
	commit, err := batch.Commit(context.Background())
	return commit, err
}

func partition(bigList []string, partitionSize int) [][]string {

	collectionLen := len(bigList)
	numFullPartitions := collectionLen / partitionSize
	capacity := numFullPartitions
	if collectionLen%partitionSize != 0 {
		capacity++
	}
	result := make([][]string, capacity)
	var i int
	for ; i < numFullPartitions; i++ {
		result[i] = bigList[i*partitionSize : (i+1)*partitionSize]
	}
	if collectionLen%partitionSize != 0 { // left over
		result[i] = bigList[i*partitionSize : collectionLen]
	}
	return result
}
