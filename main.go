package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"os"
)

type Folder struct {
	ParentID string
}

func (f Folder) String() string {
	return fmt.Sprintf("[ParentID=%s]", f.ParentID)
}

func main() {
	ctx := context.TODO()

	// Log the DB host
	if addr := os.Getenv("FIRESTORE_EMULATOR_HOST"); addr != "" {
		log.Info(addr)
	}

	// Create client
	log.Info("Creating client")
	client, err := firestore.NewClient(ctx, "irisvr-shared")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done.
	defer client.Close()

	// Start a transaction
	log.Info("Starting transaction")
	if err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		_, err := tx.Get(client.Doc("Folders/Sports"))
		if err != nil {
			return err
		}

		d := client.Doc("Folders/Sports")
		f := &Folder{ParentID: ""}
		if err := tx.Create(d, f); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatalf("Failed to execute transaction: %v", err)
	}

	log.Info("Querying")

	// Let's do a global query for all entities w/ NodeID = 1

	iter := client.Collection("Folders").Where("ParentID", "==", "sports").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		var f Folder
		if err := doc.DataTo(&f); err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		log.Info("Found folder:", f)
	}
}
