package main

import (
	"cloud.google.com/go/firestore"
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
	Key      string          `@(SimpleFieldPath | String)`
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
		`|(?P<SimpleFieldPath>[a-zA-Z_][a-zA-Z0-9_\.]*)` +
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
	limit := c.Int("limit")
	batch := c.Int("batch")
	orderby := c.String("orderby")

	client, err := createClient(credentials)
	if err != nil {
		return cliClientError(err)
	}

	parser := getParser()

	collectionRef := client.Collection(collectionPath)
	query := collectionRef.Limit(batch)

	for i := 1; i < c.NArg(); i++ {
		queryString := c.Args().Get(i)
		var parsedQuery Firestorequery
		if err := parser.ParseString(queryString, &parsedQuery); err != nil {
			return cli.NewExitError(fmt.Sprintf("Error parsing query '%s' %v", queryString, err), 83)
		}

		query = query.Where(parsedQuery.Key, parsedQuery.Operator, parsedQuery.Value.get())
	}

	// order by
	if orderby != "" {
		query = query.OrderBy(orderby, getDir(c.String("orderdir")))
	} else {
		// default ordering for batched queries must be the documentID
		query = query.OrderBy(firestore.DocumentID, firestore.Asc)
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

	var displayItems []map[string]interface{}

	// make queries with a maximum of `batch` results until we have `limit` results or no more documents are returned
	for limit > 0 {
		documentIterator := query.Documents(context.Background())

		docs, err := documentIterator.GetAll()
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to get documents. \n%v", err), 84)
		}

		var last *firestore.DocumentSnapshot
		for _, doc := range docs {
			var displayItem = make(map[string]interface{})
			last = doc

			displayItem["ID"] = doc.Ref.ID
			displayItem["CreateTime"] = doc.CreateTime
			displayItem["ReadTime"] = doc.ReadTime
			displayItem["UpdateTime"] = doc.UpdateTime
			displayItem["Data"] = doc.Data()
			displayItems = append(displayItems, displayItem)
		}

		if len(docs) == 0 {
			// no more results
			limit = 0
		} else {
			limit -= len(docs)
			// we need to figure out what the ordering is and add the correct fields
			if orderby != "" {
				query = query.StartAfter(last.Data()[orderby])
			} else {
				query = query.StartAfter(last.Ref.ID)
			}
		}
		fmt.Println(len(docs), limit)
	}

	jsonString, _ := marshallData(displayItems)
	_, err = fmt.Fprintln(c.App.Writer, jsonString)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to print result \n%v", err), 85)
	}

	return nil
}
