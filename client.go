package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"os"
)

// create client or fails.
//
// NOTE: If FIRESTORE_EMULATOR_HOST is set, it will set a
// default projectid if none has been set.

func createClient(credentials string) (*firestore.Client, error) {
	var err error
	var firebaseApp *firebase.App
	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" {
		if projectId == "" {
			projectId = "default"
		}
		client, err := firestore.NewClient(context.Background(), projectId)
		return client, err
	} else if credentials != "" {
		sa := option.WithCredentialsFile(credentials)
		config := getConfig()
		firebaseApp, err = firebase.NewApp(context.Background(), config, sa)
		if err != nil {
			return nil, err
		}
		return firebaseApp.Firestore(context.Background())
	} else {
		// This will use GOOGLE_APPLICATION_CREDENTIALS
		config := getConfig()
		firebaseApp, err = firebase.NewApp(context.Background(), config)
		if err != nil {
			return nil, err
		}
		return firebaseApp.Firestore(context.Background())
	}
}

func getConfig() *firebase.Config {
	config := firebase.Config{}
	if projectId != "" {
		config.ProjectID = projectId
	}
	return &config
}
