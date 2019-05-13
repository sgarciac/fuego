package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

// Global configuration
var credentials string

// Common errors
func cliClientError(err error) *cli.ExitError {
	return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
}

func main() {
	app := cli.NewApp()
	app.Version = "0.4.0"
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

	writingFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "timestamp, ts",
			Usage: "treat strings values that match rfc3339 as timestamps",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "collections",
			Aliases: []string{"c"},
			Usage:   "List the root level collections",
			Action:  collectionsCommandAction,
		},
		{
			Name:      "add",
			Aliases:   []string{"a"},
			Usage:     "Add a new document to a collection",
			ArgsUsage: "collection-path json-document",
			Action:    addCommandAction,
			Flags:     writingFlags,
		},
		{
			Name:      "set",
			Aliases:   []string{"s"},
			Usage:     "Set the contents of a document",
			ArgsUsage: "collection-path document-id json-document",
			Action:    setCommandAction,
			Flags:     writingFlags,
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
				cli.StringSliceFlag{
					Name:  "select",
					Usage: "Return only `FIELD_PATH` fields in result. Parameter can be given multiple times",
				},
				cli.IntFlag{
					Name:  "batch, b",
					Usage: "Will fetch documents in several requests with at most `BATCH` documents per request to prevent timeouts.",
					Value: 100,
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
