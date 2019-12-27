<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [How to contribute to Fuego](#how-to-contribute-to-fuego)
  - [Building fuego](#building-fuego)
  - [Testing](#testing)
  - [Releases](#releases)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# How to contribute to Fuego

Fuego is written in the [go programming language](https://golang.org/) (1.12 or
newer), you need to have it installed and be familiar with it, to contribute
code to fuego.

## Building fuego

Fuego uses go modules, which means you'll need at least go version 1.11 to
compile the program.

Fork and clone the repository and then build it:

```sh
cd fuego
go build
```

All dependencies will be downloaded and built automatically and the fuego binary
will be created in the project directory (use _go install_ to move the binary to
your GOPATH if you wish).

## Testing

The tests are located in [test](./test/test) as a shuint2 shell script. 
To run the tests, you'll need to install:

 * [firestore emulator](https://firebase.google.com/docs/rules/emulator-setup) (via either gcloud or firebase CLI)
 * [jq](https://stedolan.github.io/jq/)
 * uuidgen (apt-get install uuidgen-runtime in gnu/linux)

To execute the tests, follow these steps 

1. Run the [firestore emulator](https://firebase.google.com/docs/rules/emulator-setup)

2. In a different shell, go into the `tests` directory and run the tests script:

```
cd tests
./tests
or
./tests -- nameoffirsttest nameofsecondtest
```

## Releases

Releases are managed by goreleaser.

Steps:

1. export GITHUB_TOKEN=`your_token`
2. Update version in main.go (i.e v0.1.0)
3. Update CHANGELOG.md (set version)
4. Commit changes
5. git tag -a v0.1.0 -m "First release"
6. git push origin v0.1.0
7. goreleaser (options: --rm-dist --release-notes=<file>)
