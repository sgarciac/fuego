package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"log"
	"strings"
	"sync"
)

type PathType int

const (
	DocumentPath   PathType = 0
	CollectionPath PathType = 1
)

func (p PathType) String() string {
	switch p {
	case DocumentPath: return "Document"
	case CollectionPath: return "Collection"
	default:         return "UNKNOWN"
	}
}

type CopyOption struct {
	merge bool
	overwrite bool
}

func pathType(p string) PathType {
	return PathType(len(strings.Split(strings.Trim(p, "/"), "/")) % 2)
}

func copyCommandAction(c *cli.Context) error {
	argsLength := len(c.Args())

	if argsLength != 2 {
		return cli.NewExitError("Wrong number of arguments", 85)
	}

	sourceCollectionOrDocumentPath := strings.Trim(c.Args().Get(0), "/")
	targetCollectionOrDocumentPath := strings.Trim(c.Args().Get(1), "/")

	merge := c.Bool("merge")
	overwrite := c.Bool("overwrite")

	sc := c.String("src-credentials")
	dc := c.String("dest-credentials")

	if sc == "" {
		sc = credentials
	}

	if dc == "" {
		dc = credentials
	}

	option := CopyOption{
		merge:    merge,
		overwrite: overwrite,
	}

	sType := pathType(sourceCollectionOrDocumentPath)
	tType := pathType(targetCollectionOrDocumentPath)

	if sType != tType {
		return cli.NewExitError(fmt.Sprintf("Can't copy from %s to %s", sType.String(), tType.String()), 87)
	}

	sourceClient, err := createClient(sc)
	if err != nil {
		return cliClientError(err)
	}

	targetClient, err := createClient(dc)
	if err != nil {
		return cliClientError(err)
	}

	if sType == CollectionPath {
		log.Println("copying collection")
		copyCollection(
			sourceClient.Collection(sourceCollectionOrDocumentPath),
			targetClient.Collection(targetCollectionOrDocumentPath),
			option,
			)
	}

	if sType == DocumentPath {
		log.Println("copying document")
		copyDocument(
			sourceClient.Doc(sourceCollectionOrDocumentPath),
			targetClient.Doc(targetCollectionOrDocumentPath),
			option,
			)
	}

	log.Println("Done")

	return nil
}

func copyDocument(source, target *firestore.DocumentRef, option CopyOption)  {
	targetDoc, err := target.Get(context.Background())

	if targetDoc == nil {
		if err != nil {
			log.Fatalf("get document %s error: %s\n", target.Path, err)
			return
		}
		return
	}

	if targetDoc.Exists() && !option.overwrite {
		log.Fatalf("skipped document %s, because it already exists. use --override to override \n", target.Path)
		return
	}

	sourceDoc, err := source.Get(context.Background())

	if sourceDoc == nil {
		if err != nil {
			log.Fatalf("get document %s error: %s\n", source.Path, err)
			return
		}
		return
	}

	var options []firestore.SetOption
	if option.merge {
		options = append(options, firestore.MergeAll)
	}

	if sourceDoc.Exists() {
		_, err := target.Set(context.Background(), sourceDoc.Data(), options...)
		if err != nil {
			log.Fatalf("copy document %s to %s error: %s\n", source.Path, target.Path, err)
			return
		}
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	collections := source.Collections(context.Background())

	for {
		next, err := collections.Next()
		if err == iterator.Done {
			return
		}
		if err != nil {
			log.Fatal(err)
			return
		}

		wg.Add(1)
		go func(s, t *firestore.CollectionRef, wwg *sync.WaitGroup) {
			defer wwg.Done()
			copyCollection(s, t, option)
		}(next, target.Collection(next.ID), &wg)
	}
}

func copyCollection(source, target *firestore.CollectionRef, option CopyOption)  {
	refs := source.DocumentRefs(context.Background())
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		next, err := refs.Next()
		if err == iterator.Done {
			return
		}
		if err != nil {
			log.Fatal(err)
			return
		}
		wg.Add(1)
		go func(s, t *firestore.DocumentRef, wwg *sync.WaitGroup) {
			defer wwg.Done()
			copyDocument(s, t, option)
		}(next, target.Doc(next.ID), &wg)
	}
}