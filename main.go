package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Global configuration
var credentials string

// Common errors
func cliClientError(err error) *cli.ExitError {
	return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
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

func marshallData(object interface{}) (string, error) {
	buffer, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
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
			Usage:     "Add a new document to a collection",
			ArgsUsage: "collection-path json-document",
			Action:    addCommandAction,
		},
		{
			Name:      "set",
			Aliases:   []string{"s"},
			Usage:     "Set the contents of a document",
			ArgsUsage: "collection-path document-id json-document",
			Action:    setCommandAction,
		},
		{
			Name:      "get",
			Aliases:   []string{"g"},
			Usage:     "Get a document from a collection",
			ArgsUsage: "collection-path document-id",
			Action:    getCommandAction,
		},
		{
			Name:      "query",
			Aliases:   []string{"q"},
			Usage:     "Query a collection",
			ArgsUsage: "collection-path QUERY*",
			Action:    queryCommandAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "orderby, ob",
					Usage: "`FIELD_PATH` to order results by",
				},
				cli.StringFlag{
					Name:  "orderdir, od",
					Usage: "`DIRECTION` to order results (options: ASC/DESC)",
					Value: "DESC",
				},
				cli.IntFlag{
					Name:  "limit, l",
					Usage: "Fetch a maximum of `LIMIT` documents",
					Value: 100,
				},
				cli.StringFlag{
					Name:  "startat, sat",
					Usage: "Results start at document `ID`",
				},
				cli.StringFlag{
					Name:  "startafter, sar",
					Usage: "Results start after document `ID`",
				},
				cli.StringFlag{
					Name:  "endat, ea",
					Usage: "Results end at document `ID`",
				},
				cli.StringFlag{
					Name:  "endbefore, eb",
					Usage: "Results end before document `ID`",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
