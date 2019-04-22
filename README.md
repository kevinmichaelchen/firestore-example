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

## To subcollection or not to subcollection
```
INFO[0006] Iterated over 52000 / 107251 (48.48%) sport roots in 2.266097114s 
INFO[0009] Iterated over 52802 / 52802 (100.00%) sport subs in 1.61091644s 
INFO[0011] Iterated over 50 / 107251 (0.05%) food roots in 1.653059588s 
INFO[0011] Iterated over 25 / 25 (100.00%) food subs in 274.553941ms
```

We have:
- 107251 documents in "folders", and of those
  - 50 are foods
  - 52000 are sports
  - 55201 are random docs
- 52802 documents in "folders/sports/folders"
- 25 documents in "folders/foods/folders"

You'd think querying for 50 food roots would be quick since the result set
isn't very large, but it ends up taking ~1.6s.

Querying for a small subcollection, on the other hand, is blazing fast: 274ms.