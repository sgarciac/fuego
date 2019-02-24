package main

import (
	"encoding/json"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Global configuration
var credentials string

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

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "Fuego"
	app.Usage = "A firestore client"
	app.EnableBashCompletion = true

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
			Action:    addCommandAction,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
