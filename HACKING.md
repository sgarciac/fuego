
# How to contribute to Fuego

Fuego is written in the [go programming language](https://golang.org/) (1.12 or
newer), you need to have it installed and be familiar with it, to contribute
code to fuego.

All new features should be based on the ```develop``` branch. The ```master```
branch always contains the current stable version and the latest release.

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

The tests are located in [test](./test/test) as a shunit2 shell script, against
a the firebase emulator.

To run the tests, you'll need to install:

 * [firestore emulator](https://firebase.google.com/docs/rules/emulator-setup) (via either gcloud or firebase CLI)
 * [jq](https://stedolan.github.io/jq/)
 * uuidgen (apt-get install uuidgen-runtime in gnu/linux)

To execute the tests, follow these steps 

1. Run the [firestore emulator](https://firebase.google.com/docs/rules/emulator-setup)
```
$ gcloud beta emulators firestore start --host-port localhost:8080
Executing: /home/user/google-cloud-sdk/platform/cloud-firestore-emulator/cloud_firestore_emulator start --host=localhost --port=8080
[firestore] API endpoint: http://localhost:8080
[firestore] If you are using a library that supports the FIRESTORE_EMULATOR_HOST environment variable, run:
[firestore] 
[firestore]    export FIRESTORE_EMULATOR_HOST=localhost:8080
[firestore] 
[firestore] Dev App Server is now running.
```

2. In a different shell, go into the `tests` directory and run the tests script:

```
cd tests
./tests
or
./tests -- nameoffirsttest nameofsecondtest
```
## Branches (work in progress)

```master``` branch always contains the latest stable version of fuego, and
corresponds to the latest release. Under normal circumstances, only the
```develop``` branch should be merged into master. (Exceptions being hot fixes).


```develop``` branch contains the current development version. It should be kept
as clean as possible. It should always compile and pass the tests. Most new
development should start from this branch.

Other branches should be named as following:

```group/name```

Where group is one of the following items:

  * *bug* : for bug fixes.
  * *feat* : for new features.
  * *doc* : for documentation improvements.
  * *chore* : for miscelaneous changes, refactors, etc.
  
Example: ```feat/add-xyz-command```.  
  
The name of the branch should be short, hyphen-separated and represent the
purpose of the branch. 

Branches should be created with a single purpose and have a reduced number of
commits.  Use 'smash and merge' if possible when merging to develop. Keep the
develop branch free of 'wip's commits.


## Releases

Releases are managed by a github action and a new release is created whenever 
something is pushed into the ```master``` branch. Therefore pushing into master
should be done carefully. 

Release notes are taken from files in the ```release-notes``` directory,
following this convention: 

release-notes/*version*.md  (i.e. release-notes/0.15.0.md)

In order to create a new release you should:

1. Change the version number in the main.go file.
2. Create a relese-notes/*version*.md file that contains the release notes.
3. Update the CHANGELOG.md file.
4. Create a PR from develop to master
5. Admin will merge the PR.

### Creating a release locally.

Releases are managed by goreleaser.

Steps:

1. export GITHUB_TOKEN=`your_token`
2. Update version in main.go (i.e v0.1.0)
3. Update CHANGELOG.md (set version)
4. Commit changes
5. git tag -a v0.1.0 -m "First release"
6. git push origin v0.1.0
7. goreleaser (options: --rm-dist --release-notes=<file>)
