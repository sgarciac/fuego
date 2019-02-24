package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
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

	client, err := firebaseApp.Firestore(context.Background())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// unmarshall data
func unmarshallData(data string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(data)
	var buffer []byte
	if strings.HasPrefix(trimmed, "{") {
		buffer = []byte(trimmed)
	} else {
		var err error
		buffer, err = ioutil.ReadFile(data)
		if err != nil {
			return nil, err
		}
	}
	var object map[string]interface{}
	err := json.Unmarshal(buffer, &object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func addData(
	client *firestore.Client,
	collection string,
	data string) (string, error) {

	object, err := unmarshallData(data)
	if err != nil {
		return "", err
	}

	doc, _, err := client.
		Collection(collection).
		Add(context.Background(), object)

	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "Fuego"
	app.Usage = "A firestore client"
	app.EnableBashCompletion = true

	var credentials string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "credentials, c",
			Destination: &credentials,
			Usage:       "Load google application credentials from `FILE`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "add",
			Aliases:   []string{"a"},
			Usage:     "Add a new document",
			ArgsUsage: "collection-path json",
			Action: func(c *cli.Context) error {
				collectionPath := c.Args().First()
				client, err := createClient(credentials)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
				}
				id, err := addData(client, collectionPath, c.Args().Get(1))
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("Failed to add data. \n%v", err), 80)
				}
				fmt.Fprintf(c.App.Writer, "%v\n", c.Args().Get(1))
				fmt.Fprintf(c.App.Writer, "%v\n", id)

				defer client.Close()

				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
