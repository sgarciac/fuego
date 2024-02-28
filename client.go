package main

import (
	"context"
	"os"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// create client or fails.
//
// NOTE: If FIRESTORE_EMULATOR_HOST is set, it will set a
// default projectid if none has been set.

func createClient(credentials string) (*firestore.Client, error) {
	return createClientWithProjectId(credentials, projectId)
}

func createClientWithProjectId(credentials string, projectId string) (*firestore.Client, error) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" {
		if projectId == "" {
			projectId = "default"
		}
		return firestore.NewClient(context.Background(), projectId)
	}

	if database == "" {
		database = firestore.DefaultDatabaseID
	}

	options := make([]option.ClientOption, 0)
	if credentials != "" {
		options = append(options, option.WithCredentialsFile(credentials))
	}

	return firestore.NewClientWithDatabase(context.Background(), projectId, database, options...)
}

func getConfigWithProjectId(projectId string) *firebase.Config {
	config := firebase.Config{}
	if projectId != "" {
		config.ProjectID = projectId
	}
	return &config
}
