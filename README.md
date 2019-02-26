# fuego
A command-line firestore client

## Installation

Install fuego via [go](https://golang.org/dl/):

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

1. Use the ```--credentials``` flag, i.e.:

```sh
fuego --credentials ./my-account-service-private-key.json get people Rv7ZfnLQWprdXuulqMdf
```

or

2. Via the GOOGLE_APPLICATION_CREDENTIALS environment variable:

```sh
export GOOGLE_APPLICATION_CREDENTIALS=./my-account-service-private-key.json
fuego get people Rv7ZfnLQWprdXuulqMdf
```

### Writing and reading data

You can add new documents:

```sh
fuego add people '{"name": "sergio", "age": 41}'
# Rv7ZfnLQWprdXuulqMdf
```

Of fetch them:

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

You can also order the results using the ```--orderby``` and ```--orderdir``` flags. For example, to sort our nobel laureates by country of origin:

```sh
fuego query --orderby country
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
fuego query --limit 1 --orderby year --orderdir ASC nobel "birthplace.country == USA" 
```

oops, we get the following error from the server:

```
rpc error: code = FailedPrecondition desc = The query requires an index. 
You can create it here: 
https://console.firebase.google.com/project/myproject/database/firestore/indexes?create_index=EgVub2JlbBoWChJiaXJ0aH....
```

After creating the index, we re-run the query and obtain:

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

#### Pagination of query results

You can use the firestore pagination parameters, combining --limit with the flags --startat, --startafter, --endat, and --endbefore, which all accept the ID of a document.

### Hacking

Creating binary executables:

```sh
(gox -os="linux darwin windows" -arch="amd64" -output="dist/fuego_{{.OS}}_{{.Arch}}")
(cd dist; gzip *)

```

Releasing on github:

```sh
# create and push a tag:
# git tag -a v0.0.1 -m "Release description"
# git push --tags
export GITHUB_TOKEN=mytoken
export TAG=v0.0.1
ghr -t $GITHUB_TOKEN -u processone --replace --draft  $TAG dist/
```
