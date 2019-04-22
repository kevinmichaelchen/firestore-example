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

## FAQ

## Is reading after writing in a transaction okay?
No. You literally cannot call tx.Get after doing a write operation in a transaction.
If you do, you'll get this error: `firestore: read after write in transaction`.

## Will tx.Create return an error if the path already exists?
Yes. Firestore has a built-in tx.Create function 
which will error if the document already exists.

Such a thing was possible with Datastore, we just couldn't get it working.
This would close: https://github.com/IrisVR/citadel/issues/178

## Should we use subcollections?
It's ultimately a performance question.
Is querying a root level collection slower than querying for children of a subcollection?

Looks like it is.

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

## If we use subcollections, how do we move a folder?
Per the docs:
> You do not need to "create" or "delete" collections. 
After you create the first document in a collection, 
the collection exists. If you delete all of the documents 
in a collection, it no longer exists.

So to move a folder, we just need to recursively clone all descendant documents
to their new paths, and then delete the old documents pointed to by the old paths.

## Can entities support a string hashtable for metadata?
Yes. Unlike Datastore entities, Firestore documents can have a 
`map[string]string` field for arbitrary metadata.

## Do collection names have to be globally unique?
No. The following path is valid: `folders/F1/folders/F2`.
The names of documents within collections should be unique, however.

## Can we query for a folder's children inside of a transaction?
If we use subcollections, we won't need to query for children.
They will just be available in the subcollection.

## What is used for concurrency control?
Datastore used [optimistic concurrency control](https://en.wikipedia.org/wiki/Optimistic_concurrency_control).
If two transactions were simultaneously occurring, the first to commit would win.
When the second transaction tried to commit, it would fail since the timestamps 
on the entities were changed by the first transaction.

## Does delete throw an error if the document is not found?
No. The following code does not throw an error:
```golang
if _, err := client.Doc("folders/bullshit").Delete(ctx); err != nil {
  log.Fatalf("could not delete document: %v", err)
}
```

## Is there a 500 entity limit?