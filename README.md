<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [fuego](#fuego)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Authentication](#authentication)
    - [List collections](#list-collections)
    - [Writing and reading data](#writing-and-reading-data)
      - [A note on types](#a-note-on-types)
    - [Queries](#queries)
      - [Value and field path types in filters](#value-and-field-path-types-in-filters)
      - [Selecting specific fields](#selecting-specific-fields)
      - [Pagination of query results](#pagination-of-query-results)
  - [Hacking](#hacking)
    - [Testing](#testing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# fuego

[![Mentioned in Awesome Firebase](https://awesome.re/mentioned-badge.svg)](https://github.com/jthegedus/awesome-firebase)

A command-line firestore client

## Installation

Install fuego via [go get](https://golang.org/dl/):

```sh
go get github.com/sgarciac/fuego
```

Or use one of the precompiled binaries (untested) from the [latest release](https://github.com/sgarciac/fuego/releases).

## Usage

### Authentication

You'll need a Service Account key file to be able to access your project's
firestore database. To create a service account private key file, if you don't
have one, go to your firebase project console, then _Project settings_ and then
click on the _Service accounts_ tab and generate a new private key.

Once you have your service account key file, fuego will be able to find it using
one of the following options:

1. Use the ```--credentials``` flag everytime you execute fuego, i.e.:

```sh
fuego --credentials ./my-account-service-private-key.json get mycollection mydocumentid
```

or

2. Via the GOOGLE_APPLICATION_CREDENTIALS environment variable:

```sh
export GOOGLE_APPLICATION_CREDENTIALS=./my-account-service-private-key.json
fuego get mycollection mydocumentid
```

### List collections

```sh
fuego collections
```

Will return the list of projet's collections.

### Writing and reading data

You can add new documents, using JSON:

```sh
fuego add people '{"name": "sergio", "age": 41}'
# Rv7ZfnLQWprdXuulqMdf <- fuego prints the ID of the newly created document
```

Of fetch them, using the ID:

```sh
fuego get people Rv7ZfnLQWprdXuulqMdf
# {
#    "age": 41,
#    "name": "sergio"
# }
```

You can also update an existing document:

```
fuego set people Rv7ZfnLQWprdXuulqMdf '{"name": "sergio", "age": 42}' # It's my birthday!
```

In both ```add``` and ```set``` commands, the document argument can be either a
json string (if it starts with the character '{') or a path to a json file, i.e.:

```sh
fuego add animals ./dog.json
```

#### A note on types

fuego read and write commands are constrained by JSON data types: string,
number, object, array and boolean, which don't cover all of firestore data
types. 

When writing data, you can make fuego treat all strings that match the
rfc3339 datetime format as firestore timestamps, using the --timestamp (or --ts) flag. For
example:

```sh
fuego add --ts dates '{"randomdate": "2012-11-01T22:08:41+00:00"}'
```

will add a new document whose field named "randomdate" is a timestamp and not a string.

### Queries

Let's explain queries by example. First, we'll create a collection of physics
nobel laureates, 

```sh
fuego add nobel '{"name": "Arthur Ashkin", "year": 2018, "birthplace": {"country":"USA", "city": "New York"}}'
fuego add nobel '{"name": "Gerard Mourou", "year": 2018, "birthplace": {"country":"FRA", "city": "Albertville"}}'
fuego add nobel '{"name": "Donna Strickland", "year": 2018, "birthplace": {"country":"CAN", "city": "Guelph"}}'
fuego add nobel '{"name": "Rainer Weiss", "year": 2017, "birthplace": {"country":"DEU", "city": "Berlin"}}'
fuego add nobel '{"name": "Kip Thorne", "year": 2017, "birthplace": {"country":"USA", "city": "Logan"}}'
fuego add nobel '{"name": "Barry Barish", "year": 2017, "birthplace": {"country":"USA", "city": "Omaha"}}'
fuego add nobel '{"name": "David Thouless", "year": 2016, "birthplace": {"country":"GBR", "city": "Bearsden"}}'
```

We can query the full collection:

```sh
fuego query nobel
# Prints all our nobel laureates like this:
# [
#    {
#        "CreateTime": "2019-02-26T02:39:45.293936Z",
#        "Data": {
#            "birthplace": {
#                "city": "Bearsden",
#                "country": "GBR"
#            },
#            "name": "David Thouless",
#            "year": 2016
#        },
#        "ID": "BJseSVoBatOOt8gcwZWx",
#        "ReadTime": "2019-02-26T02:55:19.419627Z",
#        "UpdateTime": "2019-02-26T02:39:45.293936Z"
#    },
# .... etc
```

Which will fetch and display the documents in the collection, unfiltered. By default, fuego will fetch only 100 documents. You can change the limit using the ```--limit``` flag.

You can also order the results using the ```--orderby``` and ```--orderdir```
flags. For example, to sort our nobel laureates by country of origin, in
ascending order:

```sh
fuego query --orderby birthplace.country --orderdir ASC nobel
``` 

You can add filters, using the firestore supported operators (>, <, >=, <= and ==). You can combine several filters in a single query. For example, to get the 2018 nobel laureates from the USA:

```sh
fuego query nobel 'birthplace.country == "USA"' 'year == 2018'
```

which will print:

```json
[
    {
        "CreateTime": "2019-02-26T02:14:02.692077Z",
        "Data": {
            "birthplace": {
                "city": "New York",
                "country": "USA"
            },
            "name": "Arthur Ashkin",
            "year": 2018
        },
        "ID": "glHCUu7EZ3gkuDaVlXqv",
        "ReadTime": "2019-02-26T03:00:15.576398Z",
        "UpdateTime": "2019-02-26T02:59:55.889775Z"
    }
]

```

Let's say we want to find the least recent nobel from the USA, we can write the following query:

```sh
fuego query --limit 1 --orderby year --orderdir ASC nobel "birthplace.country == 'USA'" 
```

oops, we get the following error from the server, because our query needs an index to work:

```
rpc error: code = FailedPrecondition desc = The query requires an index. 
You can create it here: 
https://console.firebase.google.com/project/myproject/database/firestore/indexes?create_index=EgVub2JlbBoWChJiaXJ0aH....
```

After creating the index, we re-run the query and now we obtain:

```json
[
    {
        "CreateTime": "2019-02-26T02:39:44.458647Z",
        "Data": {
            "birthplace": {
                "city": "Omaha",
                "country": "USA"
            },
            "name": "Barry Barish",
            "year": 2017
        },
        "ID": "ainH3nkOA2xusEBON2An",
        "ReadTime": "2019-02-26T03:12:07.156643Z",
        "UpdateTime": "2019-02-26T02:39:44.458647Z"
    }
]
```
#### Value and field path types in filters

I our previous examples, all the segments of the path part of a filter contained
alphanumeric or the _ character and did not start with a number. When this
conditions are met, they can be written unquoted. Otherwise, they need to be
unquoted.

```sh
fuego query weirdcollection 'really."    ".strage." but valid ".fieldname == "even blank keys are valid"'
```

As for values, numeric, string, boolean and timestamp values are supported in
filters. Examples of queries:

"age >= 34", "name == 'paul'", "married == true", and "birthday == 1977-06-28T04:00:00Z"

Note that timestamps values should use the RFC3339 format and should not be
quoted. Boolean values are represented by the unquoted *true* and *false* strings.

#### Selecting specific fields

Use the --select flag to explicitely ask for the specific fields you want to be
retrieved (you can define many using several --select)

```sh
fuego query --select name --select year --limit 1 --orderby year --orderdir ASC nobel "birthplace.country == 'USA'" 
```

#### Pagination of query results
There are two ways to page through query results. 

First you can use the firestore pagination parameters to manually page through results. 
Combining --limit with the flags --startat, --startafter, --endat, and --endbefore, 
which all accept the ID of a document.

Second you can use the --batch parameter. This will cause fuego to do the pagination
internally. This is helpful for very big queries which hit the firestore query timeout (about a minute).
Very likely you will have to increase the --limit parameter from its default.

## Hacking

Releases are managed by goreleaser.

Steps:

1. export GITHUB_TOKEN=`your_token`
2. Update version in main.go (i.e v0.1.0)
3. Update CHANGELOG.md (set version)
4. Commit changes
5. git tag -a v0.1.0 -m "First release"
6. git push origin v0.1.0
7. goreleaser (options: --rm-dist --release-notes=<file>)

### Testing
The tests are located in [test](./test/test) as a shuint2 shell script. It is using the firestore emulator provided
by gcloud. You also need to have the commandline tool [jq](https://stedolan.github.io/jq/) installed.

To execute the tests go into the `tests` directory and run:
```
./tests
or
./tests -- nameoffirsttest nameofsecondtest
```



