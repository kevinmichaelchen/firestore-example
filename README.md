## How to run

### In one tab, run
```bash
gcloud beta emulators firestore start --host-port=localhost:8080
```

### In another tab, run
```bash
export FIRESTORE_EMULATOR_HOST=localhost:8080
go run *.go
``` 

## Differences from Datastore
- You cannot call tx.Get after doing a write operation, lest you get `firestore: read after write in transaction`
- Firestore has a built-in tx.Create function which will error if the document already exists.
- Datastore

Things to test:
- entities can have arbitrary "map" / "metadata" fields
- can collection + doc names repeat? e.g., `Folders/F1/Folders/F2`
- can we do queries inside transactions? e.g., all children of a folder
- is there a 500 entity limit?
- can we literally not do an Update after a Read, or is it just recommend we not?
- are doc refs (path names) limited in number of chars?