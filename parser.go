package main

import (
	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
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

type Firestorefieldpath struct {
	Key []string `@(SimpleFieldPath | String)(Dot @(SimpleFieldPath | String))*`
}

type Firestorequery struct {
	// Must take care of the "in" operator.
	Key      []string        `@(SimpleFieldPath | String | "in" | "array-contains-any" | "array-contains")(Dot @(SimpleFieldPath | String | "in" | "array-contains-any" | "array-contains"))*`
	Operator string          `@Operator`
	Value    *Firestorevalue `@@`
}

type Firestorevalue struct {
	String   *string           `  @String`
	Number   *float64          `| @Number`
	DateTime *DateTime         `| @DateTime`
	Boolean  *Boolean          `| @("true" | "false" | "TRUE" | "FALSE")`
	List     []*Firestorevalue `| "[" {@@} "]"`
}

func (value *Firestorevalue) get() interface{} {
	if value.String != nil {
		return *value.String
	} else if value.Number != nil {
		return *value.Number
	} else if value.DateTime != nil {
		return time.Time(*value.DateTime)
	} else if value.Boolean != nil {
		return !!*value.Boolean
	} else {
		list := []interface{}{}
		for _, item := range value.List {
			list = append(list, item.get())
		}
		return list
	}
}

func getQueryParser() *participle.Parser {
	queryLexer := lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<DateTime>` + rfc3339pattern + `)` +
		`|(?P<Operator><=|>=|<|>|==|in\s|array-contains-any\s|array-contains\s)` +
		`|(?P<SimpleFieldPath>[a-zA-Z_][a-zA-Z0-9_]*)` +
		`|(?P<Number>[-+]?\d*\.?\d+)` +
		`|(?P<OpenList>\[)` +
		`|(?P<CloseList>\])` +
		`|(?P<String>('[^']*')|("((\\")|[^"])*"))` +
		`|(?P<Dot>\.)`,
	))
	parser := participle.MustBuild(
		&Firestorequery{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
	)
	return parser
}

func getFieldPathParser() *participle.Parser {
	queryLexer := lexer.Must(lexer.Regexp(`(\s+)` +
		`|(?P<SimpleFieldPath>[a-zA-Z_][a-zA-Z0-9_]*)` +
		`|(?P<String>('[^']*')|("((\\")|[^"])*"))` +
		`|(?P<Dot>\.)`,
	))
	parser := participle.MustBuild(
		&Firestorefieldpath{},
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
	)
	return parser
}
