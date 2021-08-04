install-firebase:
	npm install -g firebase-tools

test:
	cd tests;firebase emulators:start --only firestore './tests'

build:
	go build
