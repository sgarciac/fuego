package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

// Queries grammar (It is probably overkill to use a parser generator)
type Boolean bool

// Capture a bool
func (b *Boolean) Capture(values []string) error {
	*b = strings.ToUpper(values[0]) == "TRUE"
	return nil
}

type Firestorequery struct {
	Key      string          `@Ident`
	Operator string          `@Operator`
	Value    *Firestorevalue `@@`
}

type Firestorevalue struct {
	String  *string  `  @String`
	Number  *float64 `| @Number`
	Boolean *Boolean `| @("true" | "false" | "TRUE" | "FALSE")`
}

func (value *Firestorevalue) get() interface{} {
	if value.String != nil {
		return *value.String
	} else if value.Number != nil {
		return *value.Number
	}
	return *value.Boolean
}

func getParser() *participle.Parser {
	queryLexer := lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<Ident>[a-zA-Z_][a-zA-Z0-9_\.]*)` +
		`|(?P<Number>[-+]?\d*\.?\d+)` +
		`|(?P<String>'[^']*'|"[^"]*")` +
		`|(?P<Operator><=|>=|<|>|==)`,
	))
	parser := participle.MustBuild(
		&Firestorequery{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Bool"),
	)
	return parser
}

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()
	client, err := createClient(credentials)
	parser := getParser()

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create client. \n%v", err), 80)
	}

	collectionRef := client.Collection(collectionPath)
	var documentIterator *firestore.DocumentIterator
	if c.NArg() > 1 {
		queryString := c.Args().Get(1)
		var parsedQuery Firestorequery
		if err := parser.ParseString(queryString, &parsedQuery); err != nil {
			return fmt.Errorf("Error parsing query '%s':\n%s", queryString, err)
		}
		fmt.Fprintf(c.App.Writer, "%v\n", parsedQuery.Value.get())
		query := collectionRef.Where(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
		for i := 2; i < c.NArg(); i++ {
			queryString = c.Args().Get(i)
			if err := parser.ParseString(queryString, &parsedQuery); err != nil {
				return fmt.Errorf("Error parsing query '%s':\n%s", queryString, err)
			}
			query = query.Where(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
		}
		documentIterator = query.Documents(context.Background())
	} else {
		documentIterator = collectionRef.Documents(context.Background())
	}

	docs, err := documentIterator.GetAll()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 80)
	}

	fmt.Fprintf(c.App.Writer, "%v\n", len(docs))
	return nil
}
