package main

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"gopkg.in/urfave/cli.v1"
	"strings"
	"time"
)

// Queries grammar (It is probably overkill to use a parser generator)

// Boolean is an alias for bool.
type Boolean bool

// DateTime is an alias for time.Time
type DateTime time.Time

// Capture a bool
func (b *Boolean) Capture(values []string) error {
	*b = strings.ToUpper(values[0]) == "TRUE"
	return nil
}

// Capture a timestamp.Timestamp
func (t *DateTime) Capture(values []string) error {
	ttime, _ := time.Parse(time.RFC3339, values[0])
	*t = DateTime(ttime)
	return nil
}

type Firestorequery struct {
	Key      []string        `@(SimpleFieldPath | String)(Dot @(SimpleFieldPath | String))*`
	Operator string          `@Operator`
	Value    *Firestorevalue `@@`
}

type Firestorevalue struct {
	String   *string   `  @String`
	Number   *float64  `| @Number`
	DateTime *DateTime `| @DateTime`
	Boolean  *Boolean  `| @("true" | "false" | "TRUE" | "FALSE")`
}

func (value *Firestorevalue) get() interface{} {
	if value.String != nil {
		return *value.String
	} else if value.Number != nil {
		return *value.Number
	} else if value.DateTime != nil {
		return time.Time(*value.DateTime)
	}
	return !!*value.Boolean
}

func getParser() *participle.Parser {
	queryLexer := lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<DateTime>` + rfc3339pattern + `)` +
		`|(?P<SimpleFieldPath>[a-zA-Z_][a-zA-Z0-9_]*)` +
		`|(?P<Number>[-+]?\d*\.?\d+)` +
		`|(?P<String>('[^']*')|("[^"]*"))` +
		`|(?P<Operator><=|>=|<|>|==)` +
		`|(?P<Dot>\.)`,
	))
	parser := participle.MustBuild(
		&Firestorequery{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
	)
	return parser
}

func getDir(name string) firestore.Direction {
	if name == "DESC" {
		return firestore.Desc
	}
	return firestore.Asc
}

// query collection-path query*
func queryCommandAction(c *cli.Context) error {
	collectionPath := c.Args().First()

	// pagination
	startAt := c.String("startat")
	startAfter := c.String("startafter")
	endAt := c.String("endat")
	endBefore := c.String("endbefore")
	selectFields := c.StringSlice("select")

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	parser := getParser()

	collectionRef := client.Collection(collectionPath)
	query := collectionRef.Limit(c.Int("limit"))

	for i := 1; i < c.NArg(); i++ {
		queryString := c.Args().Get(i)
		var parsedQuery Firestorequery
		if err := parser.ParseString(queryString, &parsedQuery); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing query '%s' %v", queryString, err), 83)
		}

		query = query.WherePath(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
	}

	// order by
	if c.String("orderby") != "" {
		query = query.OrderBy(c.String("orderby"), getDir(c.String("orderdir")))
	}

	if startAt != "" {
		documentRef := collectionRef.Doc(startAt)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", startAt), 83)
		}
		query = query.StartAt(docsnap)
	}

	if startAfter != "" {
		documentRef := collectionRef.Doc(startAfter)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", startAfter), 83)
		}
		query = query.StartAfter(docsnap)
	}

	if endAt != "" {
		documentRef := collectionRef.Doc(endAt)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", endAt), 83)
		}
		query = query.EndAt(docsnap)
	}

	if endBefore != "" {
		documentRef := collectionRef.Doc(endBefore)
		docsnap, err := documentRef.Get(context.Background())
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get '%s' within the collection", endBefore), 83)
		}
		query = query.EndBefore(docsnap)
	}

	if len(selectFields) > 0 {
		query = query.Select(selectFields...)
	}

	documentIterator := query.Documents(context.Background())

	docs, err := documentIterator.GetAll()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
	}

	var displayItems []map[string]interface{}
	for _, doc := range docs {
		var displayItem = make(map[string]interface{})
		displayItem["ID"] = doc.Ref.ID
		displayItem["CreateTime"] = doc.CreateTime
		displayItem["ReadTime"] = doc.ReadTime
		displayItem["UpdateTime"] = doc.UpdateTime
		displayItem["Data"] = doc.Data()
		displayItems = append(displayItems, displayItem)
	}

	jsonString, _ := marshallData(displayItems)
	fmt.Fprintln(c.App.Writer, jsonString)
	return nil
}
