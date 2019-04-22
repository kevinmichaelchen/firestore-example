package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"os"
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
	log.Infof("Creating folder: %s", f)
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

	// Start a transaction
	log.Info("Starting transaction")
	if err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		if err := createFolder(ctx, tx, client.Doc("folders/sports")); err != nil {
			return err
		}

		if err := createFolder(ctx, tx, client.Doc("folders/sports/folders/hockey")); err != nil {
			return err
		}

		if err := createFolder(ctx, tx, client.Doc("folders/sports/folders/hockey/folders/field-hockey")); err != nil {
			return err
		}

		if err := createFolder(ctx, tx, client.Doc("folders/sports/folders/hockey/folders/ice-hockey")); err != nil {
			return err
		}

		if err := createFolder(ctx, tx, client.Doc("folders/sports/folders/baseball")); err != nil {
			return err
		}

		return nil
	}, firestore.MaxAttempts(1)); err != nil {
		log.Fatalf("Failed to execute transaction: %v", err)
	}

	// Let's do a global query for all entities w/ NodeID = 1
	log.Info("Querying")
	iter := client.Collection("folders/sports/folders").Where("ParentID", "==", "sports").Documents(ctx)
	for {
		docsnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		var f Folder
		if err := docsnap.DataTo(&f); err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		log.Info("Found folder:", f)
	}
	log.Info("Done iterating.")
}
