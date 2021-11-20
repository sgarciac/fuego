package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

// Global configuration
var credentials string
var projectId string

// Common errors
func cliClientError(err error) *cli.ExitError {
	return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
}

func main() {
	app := cli.NewApp()
	app.Version = "0.21.0"
	app.Name = "Fuego"
	app.Usage = "A firestore client"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "credentials, c",
			Destination: &credentials,
			Usage:       "Load google application credentials from `FILE`",
		},
		cli.StringFlag{
			Name:        "projectid, p",
			Destination: &projectId,
			Usage:       "Overwrite project id",
		},
	}

	displayFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "extendedjson, ej",
			Usage: "Display documents as extended json",
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
		},
		{
			Name:      "set",
			Aliases:   []string{"s"},
			Usage:     "Set the contents of a document",
			ArgsUsage: "[collection-path document-id json-document | document-path json-document]",
			Action:    setCommandAction,
			Flags: []cli.Flag{cli.BoolFlag{
				Name:  "merge",
				Usage: "if set the set operation will do a update/patch",
			}},
		},
		{
			Name:      "copy",
			Aliases:   []string{"s"},
			Usage:     "copy collection or document",
			ArgsUsage: "[collection-path collection-path | document-path document-path]",
			Action:    copyCommandAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dest-credentials, dc",
					Usage: "Google application target project credentials from `FILE`",
				},
				cli.StringFlag{
					Name:  "src-credentials, sc",
					Usage: "Google application source project credentials from `FILE`",
				},
				cli.StringFlag{
					Name:  "dest-projectid, dp",
					Usage: "Target project ID",
				},
				cli.StringFlag{
					Name:  "src-projectid, sp",
					Usage: "Source project ID",
				},
				cli.BoolFlag{
					Name:  "merge",
					Usage: "if set the set operation will do a update/patch",
				},
				cli.BoolFlag{
					Name:  "overwrite",
					Usage: "overwrite the existing collection or document",
				},
			},
		},
		{
			Name:      "get",
			Aliases:   []string{"g"},
			Usage:     "Get a document from a collection",
			ArgsUsage: "collection-path [document-id document-path]",
			Action:    getCommandAction,
			Flags:     displayFlags,
		},
		{
			Name:      "getall",
			Aliases:   []string{"ga"},
			Usage:     "Get all document from a collection by providing ids ",
			ArgsUsage: "collection-path document-id1 [document-id2 ...]",
			Action:    getAllCommandAction,
			Flags:     displayFlags,
		},
		{
			Name:      "delete",
			Aliases:   []string{"d"},
			Usage:     "Delete a document from a collection",
			ArgsUsage: "[collection-path document-id | document-path]",
			Action:    deleteCommandAction,
		},
		{
			Name:      "deleteall",
			Aliases:   []string{"da"},
			Usage:     "Delete documents from a collection without transactional support",
			ArgsUsage: "collection-path document-id1 [document-id2 ...]",
			Action:    deleteAllCommandAction,
		},
		{
			Name:      "query",
			Aliases:   []string{"q"},
			Usage:     "Query a collection",
			ArgsUsage: "[collection-path | collection-id] QUERY*",
			Action:    queryCommandAction,
			Flags: append(
				displayFlags,
				[]cli.Flag{
					cli.StringSliceFlag{
						Name:  "orderby, ob",
						Usage: "`FIELD_PATH` to order results by",
					},
					cli.BoolFlag{
						Name:  "group, g",
						Usage: "perform a group query",
					},
					cli.StringSliceFlag{
						Name:  "orderdir, od",
						Usage: "`DIRECTION` to order results (options: ASC/DESC)",
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
				}...),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
