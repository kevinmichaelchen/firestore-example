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

const (
	FoodParent   = "food"
	SportsParent = "sports"
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
	var elapsed, avgIterNextDuration time.Duration

	testDeletingNonExistentDoc(ctx, client)
	testTransactionLimit(ctx, client)

	totalRoots = countItemsInCollection(ctx, client, "folders")

	elapsed, avgIterNextDuration, n = iterateOverRootCollection(client.Collection("folders").Where("ParentID", "==", SportsParent).Documents(ctx))
	log.Infof("Iterated over %d / %d (%.2f%%) sport ROOTS in %s w/ avg iter.Next() taking %s",
		n, totalRoots, 100*(float64(n)/float64(totalRoots)), elapsed, avgIterNextDuration)
	log.Info("")

	t = countItemsInCollection(ctx, client, "folders/sports/folders")
	elapsed, avgIterNextDuration, n = iterateOverSubcollection(
		client.Collection("folders/sports/folders").Documents(ctx))
	log.Infof("Iterated over %d / %d (%.2f%%) sport SUBS in %s w/ avg iter.Next() taking %s",
		n, t, 100*(float64(n)/float64(t)), elapsed, avgIterNextDuration)
	log.Info("")

	t = countItemsInCollection(ctx, client, "folders/sports/folders/hockey/folders")
	elapsed, avgIterNextDuration, n = iterateOverSubcollection(
		client.Collection("folders/sports/folders/hockey/folders").Documents(ctx))
	log.Infof("Iterated over %d / %d (%.2f%%) hockey SUBS in %s w/ avg iter.Next() taking %s",
		n, t, 100*(float64(n)/float64(t)), elapsed, avgIterNextDuration)
	log.Info("")

	elapsed, avgIterNextDuration, n = iterateOverRootCollection(
		client.Collection("folders").Where("ParentID", "==", FoodParent).Documents(ctx))
	log.Infof("Iterated over %d / %d (%.2f%%) food ROOTS in %s w/ avg iter.Next() taking %s",
		n, totalRoots, 100*(float64(n)/float64(totalRoots)), elapsed, avgIterNextDuration)
	log.Info("")

	t = countItemsInCollection(ctx, client, "folders/foods/folders")
	elapsed, avgIterNextDuration, n = iterateOverSubcollection(
		client.Collection("folders/foods/folders").Documents(ctx))
	log.Infof("Iterated over %d / %d (%.2f%%) food SUBS in %s w/ avg iter.Next() taking %s",
		n, t, 100*(float64(n)/float64(t)), elapsed, avgIterNextDuration)
	log.Info("")
}

func testDeletingNonExistentDoc(ctx context.Context, client *firestore.Client) {
	if _, err := client.Doc("folders/bullshit").Delete(ctx); err != nil {
		log.Fatalf("could not delete bullshit: %v", err)
	}
}

func testTransactionLimit(ctx context.Context, client *firestore.Client) {
	err := client.RunTransaction(ctx, func(c context.Context, tx *firestore.Transaction) error {
		for i := 0; i < 501; i++ {
			id := uuid.Must(uuid.NewRandom()).String()
			rootLevelDocumentRef := client.Doc(fmt.Sprintf("folders/%s", id))
			rootDoc := &Folder{ID: id}

			if err := tx.Set(rootLevelDocumentRef, rootDoc); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("transaction threw error: %v", err)
	}
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

func avg(durations []time.Duration) time.Duration {
	var ns int64
	for _, duration := range durations {
		ns += duration.Nanoseconds()
	}
	return time.Duration(int64(float64(ns)/float64(len(durations))))
}

func iterateOverSubcollection(iter *firestore.DocumentIterator) (time.Duration, time.Duration, int) {
	start := time.Now()
	avgIterNextDuration, count := iterate(iter)
	elapsed := time.Since(start)
	return elapsed, avgIterNextDuration, count
}

func iterateOverRootCollection(iter *firestore.DocumentIterator) (time.Duration, time.Duration, int) {
	start := time.Now()
	avgIterNextDuration, count := iterate(iter)
	elapsed := time.Since(start)
	return elapsed, avgIterNextDuration, count
}

func iterate(iter *firestore.DocumentIterator) (time.Duration, int) {
	count := 0
	var durations []time.Duration
	for {
		start := time.Now()
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		count += 1
		durations = append(durations, time.Since(start))
		//var f Folder
		//if err := docsnap.DataTo(&f); err != nil {
		//	log.Fatalf("Failed to iterate: %v", err)
		//}
		//log.Info("Found folder:", f)
	}
	return avg(durations), count
}

func seedStuff(ctx context.Context, client *firestore.Client) {
	// Create some foods
	for i := 0; i < 0; i++ {
		id := uuid.Must(uuid.NewRandom()).String()
		seedFolder(ctx, client, id, FoodParent)
	}

	// Create a ton of root-level folders
	for i := 0; i < 0; i++ {
		log.Infof("Creating random doc #%d folder in /folders", i)
		seedRandomRootFolder(ctx, client)
	}

	// Create a ton of subcollection sports
	for i := 0; i < 0; i++ {
		log.Infof("Creating sport doc #%d in subcollection", i)
		seedSubsport(ctx, client)
	}

	// Create a ton of subcollection foods
	for i := 0; i < 10; i++ {
		log.Infof("Creating food doc #%d in subcollection", i)
		seedSubfood(ctx, client)
	}

	// Create a ton of root-level collection sports
	for i := 0; i < 0; i++ {
		log.Infof("Creating sport doc #%d in root-level collection", i)
		id := uuid.Must(uuid.NewRandom()).String()
		seedFolder(ctx, client, id, SportsParent)
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

func seedRandomRootFolder(ctx context.Context, client *firestore.Client) {
	id := uuid.Must(uuid.NewRandom()).String()
	rootLevelDocumentRef := client.Doc(fmt.Sprintf("folders/%s", id))
	rootDoc := &Folder{ID: id}

	if _, err := rootLevelDocumentRef.Set(ctx, rootDoc); err != nil {
		log.Fatalf("could not seed: %v", err)
	}
}
