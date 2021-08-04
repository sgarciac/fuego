install-firebase:
	npm install -g firebase-tools

start-emulator:
	 firebase emulators:start --only firestore&

test:
	cd tests;./tests;