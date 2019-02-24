package main

import (
	//firestore "cloud.google.com/go/firestore"
	//"context"
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"gopkg.in/urfave/cli.v1"
	//	"strings"
	//	"text/scanner"
)

// Queries grammar (overkill)
type Firestorequery struct {
	Key      string          `@Ident`
	Operator string          `@Operator`
	Value    *Firestorevalue `@@`
}

type Firestorevalue struct {
	String *string  `  @String`
	Number *float64 `| @Number`
}

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	//collectionPath := c.Args().First()
	query := c.Args().Get(1)
	queryLexer := lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<Ident>[a-zA-Z_][a-zA-Z0-9_\.]*)` +
		`|(?P<Number>[-+]?\d*\.?\d+)` +
		`|(?P<String>'[^']*'|"[^"]*")` +
		`|(?P<Operator><=|>=|<|>|=)`,
	))

	var parsedQuery Firestorequery
	parser := participle.MustBuild(
		&Firestorequery{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"),

		// participle.Elide("Comment"),
		// Need to solve left recursion detection first, if possible.
		// participle.UseLookahead(),
	)
	err := parser.ParseString(query, &parsedQuery)
	if err != nil {
		return err
	}
	//var s scanner.Scanner
	//s.Init(strings.NewReader(query))
	//s.Filename = "Example"
	//for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
	//	fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	//}
	fmt.Printf("%s\n", parsedQuery.Key)
	fmt.Printf("%s\n", parsedQuery.Operator)
	//	fmt.Printf("%v\n", *(parsedQuery.Value.String))
	fmt.Printf("%v\n", *(parsedQuery.Value.Number))

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
