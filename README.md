# fuego
A command-line firestore client

## Installation

Install fuego via [go](https://golang.org/dl/):

```sh
go get github.com/sgarciac/fuego
```

Or use one of the precompiled binaries (untested) from the release.

### Hacking

Creating binary executables:

```sh
(gox -os="linux darwin windows" -arch="amd64" -output="dist/fuego_{{.OS}}_{{.Arch}}")
(cd dist; gzip *)

```

Releasing on github:

```sh
export GITHUB_TOKEN=mytoken
export TAG=v0.0.1
ghr -t $GITHUB_TOKEN -u processone -r dpk --replace --draft  $TAG dist/
```

