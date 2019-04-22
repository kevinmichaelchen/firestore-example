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

	numRootsTotal := countItemsInCollection(ctx, client, "folders")
	rootTime, numRoots := iterateOverRootCollection(ctx, client)
	log.Infof("Iterated over %d / %d (%d%%) roots in %s", numRoots, numRootsTotal, rootTime)

	numSubsTotal := countItemsInCollection(ctx, client, "folders/sports/folders")
	subTime, numSubs := iterateOverSubcollection(ctx, client)
	log.Infof("Iterated over %d / %d (%d%%) subs in %s", numSubs, numSubsTotal, subTime)
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

func iterateOverSubcollection(ctx context.Context, client *firestore.Client) (time.Duration, int) {
	start := time.Now()
	iter := client.Collection("folders/sports/folders").Documents(ctx)
	count := iterate(ctx, client, iter)
	return time.Since(start), count
}

func iterateOverRootCollection(ctx context.Context, client *firestore.Client) (time.Duration, int) {
	start := time.Now()
	iter := client.Collection("folders").Where("ParentID", "==", "sports").Documents(ctx)
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
	// Create a ton of root-level folders
	for i := 0; i < 0; i ++ {
		log.Infof("Creating random doc #%d folder in /folders", i)
		seedRootFolder(ctx, client)
	}

	// Create a ton of subcollection sports
	for i := 0; i < 2000; i ++ {
		log.Infof("Creating sport doc #%d in subcollection", i)
		seedSubsport(ctx, client)
	}

	// Create a ton of root-level collection sports
	for i := 0; i < 0; i ++ {
		log.Infof("Creating sport doc #%d in root-level collection", i)
		seedSport(ctx, client)
	}
}

func seedSubsport(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	dr := client.Doc(fmt.Sprintf("folders/sports/folders/%s", id))
	if _, err := dr.Set(ctx, &Folder{ID: id}); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}

func seedSport(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	rootLevelDocumentRef := client.Doc(fmt.Sprintf("folders/%s", id))
	rootDoc := &Folder{ID: id, ParentID: "sports"}
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