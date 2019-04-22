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
+--------------+-----------+-----------------+---------------------+-----------------+--------------+
|     NAME     | ITER SIZE | COLLECTION SIZE | SPACE ITERATED AS % | AVG ITER NEXT() |   ELAPSED    |
+--------------+-----------+-----------------+---------------------+-----------------+--------------+
| sports roots |     52000 |          112261 | 46.32%              | 40.542µs        | 2.114722955s |
| food roots   |        50 |          112261 | 0.04%               | 31.194469ms     | 1.559749325s |
| sports subs  |     52802 |           52802 | 100.00%             | 30.356µs        | 1.608963605s |
| hockey subs  |         2 |               2 | 100.00%             | 140.829635ms    | 281.674291ms |
| food subs    |       285 |             285 | 100.00%             | 1.00639ms       | 286.887297ms |
+--------------+-----------+-----------------+---------------------+-----------------+--------------+
```

We have:
- 112261 documents in "folders", and of those
  - 50 are foods
  - 52000 are sports
  - 60211 are random docs
- 52802 documents in "folders/sports/folders"
- 285 documents in "folders/foods/folders"

You'd think querying for 50 food roots would be quick since the result set
isn't very large, but it ends up taking ~2s.

Querying for a small subcollection, on the other hand, is blazing fast.

## If we use subcollections, how do we move a folder?
Per the docs:
> You do not need to "create" or "delete" collections. 
After you create the first document in a collection, 
the collection exists. If you delete all of the documents 
in a collection, it no longer exists.

So to move a folder, we just need to recursively clone all descendant documents
to their new paths, and then delete the old documents pointed to by the old paths.

### Moving permissions
If permissions are also subcollections, we'd move those over as well.

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

Per the Firestore docs:
> A function calling a transaction (transaction function) might run more than once 
if a concurrent edit affects a document that the transaction reads.

So it seems like they use concurrency control on not just documents we write,
but even on those we read!

## Does delete throw an error if the document is not found?
No.

The following code does not throw an error:
```golang
if _, err := client.Doc("folders/bullshit").Delete(ctx); err != nil {
  log.Fatalf("could not delete document: %v", err)
}
```

## Is there a 500 entity transaction limit?
Datastore infamously prevents you from modifying more than 500 entities per transaction.

Firestore does not have this limitation.

Per the docs:
> Each transaction or batch of writes can write to a maximum of 500 documents.

However, when I actually try to test this, by calling `tx.Set` 501 times,
no error occurs.  