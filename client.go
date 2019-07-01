package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// create client or fails
func createClient(credentials string) (*firestore.Client, error) {
	var err error
	var firebaseApp *firebase.App
	if credentials != "" {
		sa := option.WithCredentialsFile(credentials)
		firebaseApp, err = firebase.NewApp(context.Background(), nil, sa)
	} else {
		// Use GOOGLE_APPLICATION_CREDENTIALS
		firebaseApp, err = firebase.NewApp(context.Background(), nil)
	}

	if err != nil {
		return nil, err
	}

	return firebaseApp.Firestore(context.Background())
}
