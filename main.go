package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"os"
	"time"
)

type Folder struct {
	ID       string
	ParentID string
	Metadata map[string]string
}

func (f Folder) String() string {
	return fmt.Sprintf("[ID=%s, ParentID=%s, Metadata=%s]", f.ID, f.ParentID, f.Metadata)
}

func createFolder(ctx context.Context, tx *firestore.Transaction, dr *firestore.DocumentRef) error {
	grandParent := dr.Parent.Parent
	isRootFolder := grandParent == nil
	f := &Folder{
		ID: dr.ID,
	}
	if !isRootFolder {
		f.ParentID = grandParent.ID
	}
	//log.Infof("Creating folder: %s", f)
	err := tx.Set(dr, f)
	if err != nil {
		err = multierror.Append(fmt.Errorf("error creating folder: %s", dr.Path), err)
	}
	return err
}

func main() {
	ctx := context.TODO()

	// Log the DB host
	if addr := os.Getenv("FIRESTORE_EMULATOR_HOST"); addr != "" {
		log.Infof("Connecting to DB at %s", addr)
	}

	// Create client
	log.Info("Creating client")
	client, err := firestore.NewClient(ctx, "irisvr-shared")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Close client when done.
	defer client.Close()

	seedStuff(ctx, client)

	var totalRoots, n, t int
	var e time.Duration

	totalRoots = countItemsInCollection(ctx, client, "folders")
	e, n = iterateOverRootCollection(ctx, client, "sports")
	log.Infof("Iterated over %d / %d (%.2f%%) sport roots in %s",
		n, totalRoots, 100 * (float64(n) / float64(totalRoots)), e)

	t = countItemsInCollection(ctx, client, "folders/sports/folders")
	e, n = iterateOverSubcollection(ctx, client, "folders/sports/folders")
	log.Infof("Iterated over %d / %d (%.2f%%) sport subs in %s",
		n, t, 100 * (float64(n) / float64(t)), e)

	e, n = iterateOverRootCollection(ctx, client, "food")
	log.Infof("Iterated over %d / %d (%.2f%%) food roots in %s",
		n, totalRoots, 100 * (float64(n) / float64(totalRoots)), e)

	t = countItemsInCollection(ctx, client, "folders/foods/folders")
	e, n = iterateOverSubcollection(ctx, client, "folders/foods/folders")
	log.Infof("Iterated over %d / %d (%.2f%%) food subs in %s",
		n, t, 100 * (float64(n) / float64(t)), e)
}

func countItemsInCollection(ctx context.Context, client *firestore.Client, path string) int {
	iter := client.Collection(path).Documents(ctx)
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		count += 1
	}
	return count
}

func iterateOverSubcollection(ctx context.Context, client *firestore.Client, path string) (time.Duration, int) {
	start := time.Now()
	iter := client.Collection(path).Documents(ctx)
	count := iterate(ctx, client, iter)
	return time.Since(start), count
}

func iterateOverRootCollection(ctx context.Context, client *firestore.Client, parentID string) (time.Duration, int) {
	start := time.Now()
	iter := client.Collection("folders").Where("ParentID", "==", parentID).Documents(ctx)
	count := iterate(ctx, client, iter)
	return time.Since(start), count
}

func iterate(ctx context.Context, client *firestore.Client, iter *firestore.DocumentIterator) int {
	count := 0
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		var f Folder
		count += 1
		if err := docsnap.DataTo(&f); err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		//log.Info("Found folder:", f)
	}
	return count
}

func seedStuff(ctx context.Context, client *firestore.Client) {
	// Create some foods
	for i := 0; i < 25; i ++ {
		id := uuid.Must(uuid.NewRandom()).String()
		seedFolder(ctx, client, id, "food")
	}

	// Create a ton of root-level folders
	for i := 0; i < 100; i ++ {
		log.Infof("Creating random doc #%d folder in /folders", i)
		seedRootFolder(ctx, client)
	}

	// Create a ton of subcollection sports
	for i := 0; i < 0; i ++ {
		log.Infof("Creating sport doc #%d in subcollection", i)
		seedSubsport(ctx, client)
	}

	// Create a ton of subcollection foods
	for i := 0; i < 25; i ++ {
		log.Infof("Creating food doc #%d in subcollection", i)
		seedSubfood(ctx, client)
	}

	// Create a ton of root-level collection sports
	for i := 0; i < 0; i ++ {
		log.Infof("Creating sport doc #%d in root-level collection", i)
		id := uuid.Must(uuid.NewRandom()).String()
		seedFolder(ctx, client, id, "sports")
	}
}

func seedSubsport(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	dr := client.Doc(fmt.Sprintf("folders/sports/folders/%s", id))
	if _, err := dr.Set(ctx, &Folder{ID: id}); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}

func seedSubfood(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	dr := client.Doc(fmt.Sprintf("folders/foods/folders/%s", id))
	if _, err := dr.Set(ctx, &Folder{ID: id}); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}

func seedFolder(ctx context.Context, client *firestore.Client, id, parentID string) {
	rootLevelDocumentRef := client.Doc(fmt.Sprintf("folders/%s", id))
	rootDoc := &Folder{ID: id, ParentID: parentID}
	if _, err := rootLevelDocumentRef.Set(ctx, rootDoc); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}

func seedRootFolder(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	rootLevelDocumentRef := client.Doc(fmt.Sprintf("folders/%s", id))
	rootDoc := &Folder{ID: id}

	if _, err := rootLevelDocumentRef.Set(ctx, rootDoc); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}