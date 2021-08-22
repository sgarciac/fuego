# TYPES

Fuego uses JSON as the data exchange format with the user. 

When writing to the database, user input's JSON documents are transformed to
Firestore documents, mapping their values as follows:

| JSON          | Firestore             |
| ------------- | ----------------      |
| Null          | Null                  |
| Boolean       | Boolean               |
| Array         | Array                 |
| Map           | Map                   |
| Number        | Floating-point number |
| String        | String                |

When displaying firestore documents as JSON, firestore values are
transformed to json values, as follows:


| Firestore            | JSON                                                        |
| -------------        | ----------------                                            |
| Null                 | Null                                                        |
| Boolean              | Boolean                                                     |
| Array                | Array                                                       |
| Map                  | Map                                                         |
| Integer              | Number                                                      |
| Floting-point number | Number                                                      |
| String               | String                                                      |
| Date and time        | String                                                      |
| Geopoint             | {"latitude": number: "longitude": number}                   |
| Reference            | {"Parent": Reference or Null, "Path": string, "Id": string} |
| Bytes                | ???                                                         |

As it must be obvious, it is impossible to directly represent all firestore
types using JSON only. Some values are impossible to express, when writing to
the database, and typing information is lost, when reading from the database.

To fix this limitation, fuego supports an extended JSON format, following the
example of [MongoDB's Extended JSON format
(v2)](https://docs.mongodb.com/manual/reference/mongodb-extended-json/).

## Fuego's extended JSON format.

### Nil

```
<null>
```

### Arrays

```
<array of extended JSON elements>

```
### Boolean

```
{"$boolean": <boolean>}
```

or 

```
<boolean>
```

### Bytes

```
{"$binary": "a base 64 string representing the array of bytes"}
```

### Date and time

```
{"$date": "<RFC3339 string>"}

```

### Floating-point number

```
{"$numberDouble": <Number> }
```

or

```
<number>
```

### Geographical point

```
{ "$geopoint":
   {
      "$longitude": <number>,
      "$latitude": <number>
   }
}
```

### Integer

```
{"$numberInt": <Number> }
```

### Map

```
<Map of extended JSON elements>
```

### Reference

???

### String

```
{"$string": <String> }
```
or

```
<String>
```
