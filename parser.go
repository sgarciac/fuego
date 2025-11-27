package main

import (
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
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
	Key      []string        `@(SimpleFieldPath | String)(Dot @(SimpleFieldPath | String))*`
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

func getQueryParser() *participle.Parser[Firestorequery] {
	queryLexer := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "DateTime", Pattern: `` + rfc3339pattern + ``},
		{Name: "Operator", Pattern: `<in>|<not-in>|<array-contains-any>|<array-contains>|<=|>=|<|>|==|!=`},
		{Name: "SimpleFieldPath", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Number", Pattern: `[-+]?\d*\.?\d+`},
		{Name: "OpenList", Pattern: `\[`},
		{Name: "CloseList", Pattern: `\]`},
		{Name: "String", Pattern: `('[^']*')|("((\\")|[^"])*")`},
		{Name: "Dot", Pattern: `\.`},
	})
	parser := participle.MustBuild[Firestorequery](
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
	)
	return parser
}

func getFieldPathParser() *participle.Parser[Firestorefieldpath] {
	queryLexer := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "SimpleFieldPath", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "String", Pattern: `('[^']*')|("((\\")|[^"])*")`},
		{Name: "Dot", Pattern: `\.`},
	})
	parser := participle.MustBuild[Firestorefieldpath](
		participle.Lexer(queryLexer),
		participle.Unquote("String"),
	)
	return parser
}
