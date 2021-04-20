package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/urfave/cli"
	"google.golang.org/api/iterator"
	"log"
	"strings"
	"sync/atomic"
	"time"
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

	sp := c.String("src-projectid")
	dp := c.String("dest-projectid")

	if sc == "" {
		sc = credentials
	}

	if dc == "" {
		dc = credentials
	}

	if sp == "" {
		sp = projectId
	}

	if dp == "" {
		dp = projectId
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

	sourceClient, err := createClientWithProjectId(sc, sp)
	if err != nil {
		return cliClientError(err)
	}

	targetClient, err := createClientWithProjectId(dc, dp)
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
	client := NewCopyClient(200, option)

	client.Run(
		NewDocumentCopyJob(source, target),
		NewCollectionIterationJob(source, target),
		)
}

func copyCollection(source, target *firestore.CollectionRef, option CopyOption)  {
	client := NewCopyClient(200, option)

	client.Run(NewDocumentIterationJob(source, target))
}

type CopyClient struct {
	jobQueue []CopyJob
	workerQueue   []chan CopyJob
	jobChannel    chan CopyJob
	workerChannel chan chan CopyJob
	WorkerCount int
	Option CopyOption
}

func NewCopyClient(workerCount int, option CopyOption) CopyClient {
	return CopyClient{
		jobChannel:    make(chan CopyJob),
		workerChannel: make(chan chan CopyJob),
		WorkerCount:   workerCount,
		Option:        option,
	}
}

func (client *CopyClient) Run(seeds... CopyJob)  {
	out := make(chan []CopyJob)

	go func() {
		// submit initial jobs
		for _, j := range seeds {
			client.submitJob(j)
		}

		for {
			// read jobs from workers
			jobs := <-out
			for _, j := range jobs {
				client.submitJob(j)
			}
		}
	}()

	// create works to handle jobs
	client.createWorkers(out)

	// match workers with jobs
	done := client.scheduleJobs()

	<- done
}

func (client *CopyClient) submitJob(job CopyJob)  {
	client.jobChannel <- job
}

func (client *CopyClient) workerReady(w chan CopyJob)  {
	client.workerChannel <- w
}

func (client *CopyClient) workerInput() chan CopyJob  {
	return make(chan CopyJob)
}

func (client *CopyClient) createWorkers(workerOutput chan []CopyJob) {
	for i:=0; i < client.WorkerCount; i++ {
		in := client.workerInput()
		go func(in chan CopyJob) {
			for {
				client.workerReady(in)
				j := <-in
				w := CopyWorker{
					Overwrite: client.Option.overwrite,
					Merge:     client.Option.merge,
				}
				jobs, err := w.handleJob(j)
				if err != nil {
					continue
				}
				workerOutput <- jobs
			}
		}(in)
	}
}

func (client *CopyClient) scheduleJobs() <- chan struct{} {
	done := make(chan struct{})
	go func() {
		j := <- client.jobChannel
		client.jobQueue = append(client.jobQueue, j)

		for {
			var activeJob CopyJob
			var activeWorker chan CopyJob

			if len(client.jobQueue) > 0 && len(client.workerQueue) > 0 {
				activeJob = client.jobQueue[0]
				activeWorker = client.workerQueue[0]
			}

			select {
			case jj := <- client.jobChannel:
				client.jobQueue = append(client.jobQueue, jj)
			case w := <- client.workerChannel:
				client.workerQueue = append(client.workerQueue, w)
			case activeWorker <- activeJob:
				client.jobQueue = client.jobQueue[1:]
				client.workerQueue = client.workerQueue[1:]
			case <- time.After(time.Second * 2):
				// no more jobs
				if len(client.jobQueue) == 0 && len(client.workerQueue) == client.WorkerCount {
					done <- struct{}{}
					return
				}
			}
		}
	}()
	return done
}

var ops int32

type CopyWorker struct {
	Overwrite bool
	Merge bool
}

func (w *CopyWorker) handleJob(j CopyJob) ([]CopyJob, error)  {
	var result []CopyJob

	if j.name == "iterateCollection" {
		return w.handleCollectionIterationJob(j)
	}

	if j.name == "iterateDocument" {
		return w.handleDocumentIterationJob(j)
	}

	if j.name == "copyDocument" {
		return w.handleCopyDocumentJob(j)
	}

	return result, nil
}

func (w *CopyWorker) handleCollectionIterationJob(j CopyJob) ([]CopyJob, error) {
	var result []CopyJob

	if next, err := j.collectionIterator.Next(); err != nil {
		if err == iterator.Done {
			return result, nil
		}
		return result, err
	} else {
		result = append(result,
			j,
			NewDocumentIterationJob(next, j.targetDocumentRef.Collection(next.ID)),
		)
		return result, nil
	}
}

func (w *CopyWorker) handleDocumentIterationJob(j CopyJob) ([]CopyJob, error) {
	var result []CopyJob

	if next, err := j.documentRefIterator.Next(); err != nil {
		if err == iterator.Done {
			return result, nil
		}
		return result, err
	} else {
		result = append(result,
			j,
			NewCollectionIterationJob(next, j.targetCollectionRef.Doc(next.ID)),
			NewDocumentCopyJob(next, j.targetCollectionRef.Doc(next.ID)),
		)
		return result, nil
	}
}

func (w *CopyWorker) handleCopyDocumentJob(j CopyJob) ([]CopyJob, error) {
	var result []CopyJob

	targetDoc, err := j.targetDocumentRef.Get(context.Background())

	if targetDoc == nil {
		if err != nil {
			log.Printf("get document %s error: %s\n", j.targetDocumentRef.Path, err)
			return result, err
		}
		return result, nil
	}

	if targetDoc.Exists() && !w.Overwrite {
		log.Printf("skipped document %s, because it already exists. use --overwrite to overwrite \n", j.targetDocumentRef.Path)
		return result, nil
	}

	sourceDoc, err := j.documentRef.Get(context.Background())

	if sourceDoc == nil {
		if err != nil {
			log.Printf("get document %s error: %s\n", j.documentRef.Path, err)
			return result, err
		}
		return result, nil
	}

	var options []firestore.SetOption
	if w.Merge {
		options = append(options, firestore.MergeAll)
	}

	atomic.AddInt32(&ops, 1)

	if sourceDoc.Exists() {
		log.Printf("set document: %d, %s \n", ops, j.targetDocumentRef.Path)
		_, err = j.targetDocumentRef.Set(context.Background(), sourceDoc.Data(), options...)
		if err != nil {
			log.Printf("copy document %s to %s error: %s\n", j.documentRef.Path, j.targetDocumentRef.Path, err)
			return result, err
		}
	}

	return result, nil
}

type CopyJob struct {
	// iterateCollection iterateDocument copyDocument
	name string

	documentRef *firestore.DocumentRef

	collectionIterator *firestore.CollectionIterator
	targetDocumentRef *firestore.DocumentRef

	documentRefIterator *firestore.DocumentRefIterator
	targetCollectionRef *firestore.CollectionRef
}

func NewCollectionIterationJob(documentRef *firestore.DocumentRef, targetDocumentRef *firestore.DocumentRef) CopyJob {
	return CopyJob{
		name:                "iterateCollection",
		collectionIterator:  documentRef.Collections(context.Background()),
		targetDocumentRef:   targetDocumentRef,
	}
}


func NewDocumentIterationJob(collectionRef *firestore.CollectionRef, targetCollectionRef *firestore.CollectionRef) CopyJob {
	return CopyJob{
		name:                "iterateDocument",
		documentRefIterator: collectionRef.DocumentRefs(context.Background()),
		targetCollectionRef: targetCollectionRef,
	}
}

func NewDocumentCopyJob(documentRef *firestore.DocumentRef, targetDocumentRef *firestore.DocumentRef) CopyJob {
	return CopyJob{
		name:                "copyDocument",
		documentRef:         documentRef,
		targetDocumentRef:   targetDocumentRef,
	}
}