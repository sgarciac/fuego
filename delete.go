package main

import (
	"context"
	"fmt"
	"log"

	firestore "cloud.google.com/go/firestore"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
)

func deleteCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())
	var err error

	if argsLength < 1 || argsLength > 2 {
		return cli.NewExitError("Wrong number of arguments", 82)
	}

	deleteRecursive := c.Bool("recursive")
	deleteField := c.String("field")

	if deleteRecursive && deleteField != "" {
		return cli.NewExitError("recursive delete and field delete can't be combined!", 82)
	}

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	var documentRef *firestore.DocumentRef

	if argsLength == 2 {
		collectionPath := c.Args().First()
		id := c.Args().Get(1)
		collectionRef := client.Collection(collectionPath)
		documentRef = collectionRef.Doc(id)
	} else {
		documentPath := c.Args().First()
		documentRef = client.Doc(documentPath)
	}

	defer client.Close()

	if deleteRecursive {
		deleteBatch := client.Batch()
		err = deleteSubCollections(documentRef, client, deleteBatch)
		if err != nil {
			return err
		}

		_, err = deleteBatch.Commit(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("failed to delete sub-collections batch. %v\n", err), 82)
		}
	}

	if deleteField != "" {

		_, err := documentRef.Update(
			context.Background(),
			[]firestore.Update{{
				Path:  deleteField,
				Value: firestore.Delete,
			},
			})

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("failed to delete field. \n%v", err), 82)
		}

		return nil
	}

	res, err := documentRef.Delete(context.Background())
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to delete data. \n%v", res), 82)
	}
	defer client.Close()
	return nil
}

func deleteSubCollections(r *firestore.DocumentRef, c *firestore.Client, b *firestore.WriteBatch) error {

	subCollectionIter := r.Collections(context.Background())

	for {
		subCol, err := subCollectionIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("failed to iterate over sub-collections (error at %v).", subCol.Path), 82)
		}

		err = deleteCollection(subCol, c, b)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteCollection(r *firestore.CollectionRef, c *firestore.Client, b *firestore.WriteBatch) error {

	iter := r.DocumentRefs(context.Background())
	for {
		numDeleted := 0
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return cli.NewExitError(fmt.Sprintf(
					"could not iterate over sub-collection %s (error on document %s). error %s\n",
					r.ID, doc.Path, err), 82)
			}

			err = deleteSubCollections(doc, c, b)
			if err != nil {
				return err
			}

			b.Delete(doc)
			numDeleted++
		}

		if numDeleted == 0 {
			break
		}

		log.Printf("added sub-collection %s to delete-batch! (number of deleted documents: %d)", r.Path, numDeleted)
	}

	return nil
}
