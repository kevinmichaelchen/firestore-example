## How to run

### In one tab, run
```bash
gcloud beta emulators firestore start
```

### In another tab, run
```bash
export FIRESTORE_EMULATOR_HOST=WHATEVER_THE_OTHER_TAB_SAYS
go run *.go
``` 

Currently having difficulty running a transaction.
See https://github.com/GoogleCloudPlatform/golang-samples/blob/master/firestore/firestore_snippets/save.go#L312.

Things to test:
- entities can have arbitrary "map" / "metadata" fields
- can collection + doc names repeat? e.g., `Folders/F1/Folders/F2`
- can we do queries inside transactions? e.g., all children of a folder
- is there a 500 entity limit?
- can we literally not do an Update after a Read, or is it just recommend we not?
- are doc refs (path names) limited in number of chars?