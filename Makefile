install-firebase:
	npm install -g firebase-tools

test:
	cd tests;firebase emulators:exec --only firestore './tests'


build:
	go build
