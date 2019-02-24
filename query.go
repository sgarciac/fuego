package main

import (
	//firestore "cloud.google.com/go/firestore"
	//"context"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"
	"text/scanner"
)

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	//collectionPath := c.Args().First()
	query := c.Args().Get(1)
	var s scanner.Scanner
	s.Init(strings.NewReader(query))
	s.Filename = "Example"
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	}
	//client, err := createClient(credentials)
	//if err != nil {
	//	return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
	//}
	//data, err := getData(client, collectionPath, id)
	//if err != nil {
	//	return cli.NewExitError(fmt.Sprintf("Failed to get data. \n%v", err), 80)
	//}
	//fmt.Fprintf(c.App.Writer, "%v\n", data)

	//defer client.Close()
	return nil
}
