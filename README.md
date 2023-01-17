 *PLEASE READ THE [STABLE VERSION DOCUMENTATION](https://sgarciac.github.io/fuego/) IF YOU ONLY WANT TO USE FUEGO*

# fuego

[![Mentioned in Awesome Firebase](https://awesome.re/mentioned-badge.svg)](https://github.com/jthegedus/awesome-firebase)

A command-line firestore client.

## Installation

### Precompiled binaries

Download one of the precompiled binaries from the [latest
release](https://github.com/sgarciac/fuego/releases). (builds available for
windows, Linux, Macintosh/Darwin)

### Snap

Linux users can install fuego via snap. However, you'll probably need to install it using --devmode for it to be able to access
to your GOOGLE_APPLICATION_CREDENTIALS file.

```sh
snap install fuego --devmode
```

### Building it locally

If you are comfortable building programs, you can build fuego yourself using [go](https://golang.org/dl/):

```sh
git clone https://github.com/sgarciac/fuego.git
cd fuego
go build . # and 'go install .' if you want
./fuego --help
```

## Usage

### Authentication

You'll need a Service Account key file to be able to access your project's
firestore database. To create a service account private key file, if you don't
have one, go to your firebase project console, then _Project settings_ and then
click on the _Service accounts_ tab and generate a new private key.

Once you have your service account key file, fuego will be able to find it using
one of the following options:

1. Use the ```--credentials``` flag every time you execute fuego, i.e.:

```sh
fuego --credentials ./my-account-service-private-key.json get mycollection mydocumentid
```

or

2. Via the GOOGLE_APPLICATION_CREDENTIALS environment variable:

```sh
export GOOGLE_APPLICATION_CREDENTIALS=./my-account-service-private-key.json
fuego get mycollection mydocumentid
```

### Project

Firestore databases belong to projects. The google application credentials file
usually define the project that firestore will work on. You can however, if
necessary, define the project using the global option ```--projectid```.

### Firestore emulator usage

If you need to use fuego with the firestore emulator instead of a real firestore
database, set the FIRESTORE_EMULATOR_HOST environment variable to something
appropriate (usually, localhost:8080). **NOTE**: When using the emulator, you
are likely not using a GOOGLE_APPLICATION_CREDENTIALS file. Therefore, no
project will be defined. **You can set a project** using the global option
```--projectid```, otherwise it will use 'default' as the project identifier.

### List collections

```sh
fuego collections
```

Will return the list of project's collections.

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
#     "CreateTime": "2021-08-22T23:53:31.439821Z",
#     "Data": {
#         "age": 41,
#         "name": "sergio"
#     },
#     "ID": "Rv7ZfnLQWprdXuulqMdf",
#     "ReadTime": "2021-08-23T01:57:12.30626Z",
#     "UpdateTime": "2021-08-22T23:53:31.439821Z"
# }


```

Or fetch them using multiple ids:
```sh
fuego  getall people WkVlcPgEJIXzdyQS6H5d f2TbJA5DIhBfXwKrMbHP
[
# {
#     "CreateTime": "2021-08-22T23:53:31.439821Z",
#     "Data": {
#         "age": 41,
#         "name": "sergio"
#     },
#     "ID": "WkVlcPgEJIXzdyQS6H5d",
#     "ReadTime": "2021-08-23T01:57:12.30626Z",
#     "UpdateTime": "2021-08-22T23:53:31.439821Z"
# },
# {
#     "CreateTime": "2021-08-22T23:53:31.439821Z",
#     "Data": {
#         "age": 23,
#         "name": "rohan"
#     },
#     "ID": "f2TbJA5DIhBfXwKrMbHP",
#     "ReadTime": "2021-08-23T01:57:12.30626Z",
#     "UpdateTime": "2021-08-22T23:53:31.439821Z"
# }
# ]
```

You can also replace an existing document:

```
fuego set people/Rv7ZfnLQWprdXuulqMdf '{"name": "sergio", "age": 42}' # It's my birthday!
```

*Note*: we can either use the arguments ```collection-path document-id
json-data``` or ```document-path json-data```. This is also the case for the
delete and the get commands.

In both ```add``` and ```set``` commands, the document argument can be either a
JSON string (if it starts with the character ```{```) or a path to a JSON file, i.e.:

```sh
fuego add animals ./dog.json
```

To delete a document:

```sh
fuego delete people/Rv7ZfnLQWprdXuulqMdf
```

note: this won't delete any subcollection under the document.

It's also possible to delete multiple documents without transaction
```sh
fuego delete people Rv7ZfnLQWprdXuulqMdf,Rv7ZfnLQWprdXuulqMde
```


To update an existing document:

```sh
fuego set --merge people Rv7ZfnLQWprdXuulqMdf '{"location": "unknown"}'
# Rv7ZfnLQWprdXuulqMdf <- fuego prints the ID of the updated document
fuego get people Rv7ZfnLQWprdXuulqMdf

# {
#     "CreateTime": "2021-08-22T23:53:31.439821Z",
#     "Data": {
#         "age": 41,
#         "name": "sergio",
#         "location": "unknown"
#     },
#     "ID": "Rv7ZfnLQWprdXuulqMdf",
#     "ReadTime": "2021-08-23T01:57:12.30626Z",
#     "UpdateTime": "2021-08-22T23:53:31.439821Z"
# }

```

### Advanced types

Our examples here use only basic JSON to represent firestore documents. However,
JSON types are not enough to represent some of firestore types, for instance
geo-locations or timestamps.

Please read the [TYPES](TYPES.md) documentation if you want to know how fuego
maps JSON documents to firestore documents, and how to express the more advanced
type system using our 'extended JSON'.

### Subcollections

You can work on sub-collections using the full path with "/"s as separators. For
example:

```sh
fuego query countries/france/cities
```

### Queries

Let's explain queries by example. First, we'll create a collection of physics
Nobel laureates,

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
flags. For example, to sort our Nobel laureates by country of origin, in
ascending order:

```sh
fuego query --orderby birthplace.country --orderdir ASC nobel
```

You can add filters, using the firestore supported operators :

```
>, <, >=, <=, ==, !=, <in>, <not-in>, <array-contains> or <array-contains-any>
```

You can combine several filters in a
single query. For example, to get the 2018 Nobel laureates from the USA:

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


Or USA and Germany.

```sh
fuego query nobel 'birthplace.country <in> ["USA","DEU"]' 'year == 2018'
```



Let's say we want to find the least recent Nobel from the USA, we can write the following query:

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

As for values, numeric, string, boolean (true or false) and timestamp values are supported in
filters. Examples of queries:

"age >= 34", "name == 'paul'", "married == true", and "birthday == 1977-06-28T04:00:00Z"

Note that timestamps values should use the RFC3339 format and should not be
quoted. Boolean values are represented by the unquoted *true* and *false*
strings.

Arrays values should be expressed as in the following example. Notice that items
are separated by space:

```sh
fuego query cities 'name <in> ["bogota" "cali" "medellin"]'
```


#### Selecting specific fields

Use the --select flag to explicitely ask for the specific fields you want to be
retrieved (you can define many using several --select)

```sh
fuego query --select name --select year --limit 1 --orderby year --orderdir ASC nobel "birthplace.country == 'USA'"
```

#### Pagination of query results
If necessary, you can use the firestore pagination parameters to manually page through results.
Combining --limit with the flags --startat, --startafter, --endat, and --endbefore,
which all accept the ID of a document.

#### Group queries

You can make [group
queries](https://firebase.google.com/docs/firestore/query-data/queries) to query all the subcollections sharing a common ID by using
the -g flag.

```sh
fuego add cities/france/landmarks '{"name": "The Eiffel Tower"}'
fuego add cities/sf/landmarks '{"name": "Golden Gate Bridge"}'
fuego query -g landmarks
```

### Copying
Basic usage
```sh
fuego copy source target
```
#### Copying collection
We can copy a collection and its sub collections
```sh
fuego copy countries/france/cities countries/germany/cities
```
By default, existing documents in target collection will be skipped. If you want to overwrite the existing document, just use --overwrite
```sh
fuego copy countries/france/cities countries/germany/cities --overwrite
```
Also, using flag --merge let us can use merging mode to overwrite the existing documents
```sh
fuego copy countries/france/cities countries/germany/cities --overwrite --merge
```

#### Copying document
We can copy a document and its sub collections.
```sh
fuego copy countries/france countries/germany
```
Parameters --merge and --overwrite can also be used to specify the copying behavior.

#### Cross projects copying
We may have firestore in different Google projects. We can specify the source project credential by using `--src-credentials` (or `-sc`) and target project credential by using `--dest-credentials` (or `-dc`).
The default value of the `--src-credentials` and `--dest-credentials` is our current working project.
```sh
fuego copy countries/france countries/germany --src-credentials ./project-a-key.json --dest-credentials ./project-b-key.json --overwrite --merge
fuego copy countries/france/cities countries/germany/cities --src-credentials ./project-a-key.json --dest-credentials ./project-b-key.json
```
We may also have a credential that has access to different projects. We can specify the source project ID by `--src-projectid` (or `-sp`) and target project ID by using `--dest-projectid` (or `-dp`).
The default value of the `--src-prjectid` and `--dest-prjectid` is the ID of our current working project.
```sh
fuego copy countries/france countries/germany --src-projectid project-a --dest-projectid project-b --overwrite --merge
fuego copy countries/france/cities countries/germany/cities --dest-projectid prject-c
```

## Contributing

See the [HACKING](HACKING.md) file for some guidance on how to contribute.
