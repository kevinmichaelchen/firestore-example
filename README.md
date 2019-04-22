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
- Is querying a root level collection slower than querying for children of a subcollection?
- Demonstrate recursive move of a subcollection.  
- Can document have a `map[string]string` field for arbitrary metadata? Yes.
- Can collection and doc names repeat (e.g., `folders/F1/folders/F2`)? Yes.
- Can we do queries inside transactions? e.g., all children of a folder
- is there a 500 entity limit?
- optimistic concurrency control
- does delete throw an error if not found?
- can we literally not do an Update after a Read, or is it just recommend we not?
- are doc refs (path names) limited in number of chars?