package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"os"
)

type Folder struct {
	ID       string
	ParentID string
}

func (f Folder) String() string {
	return fmt.Sprintf("[ID=%s, ParentID=%s]", f.ID, f.ParentID)
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

		if err := tx.Set(
			client.Doc("folders/sports"),
			&Folder{
				ID:       "sports",
				ParentID: "",
			}); err != nil {
			return multierror.Append(
				errors.New("couldn't create sports folder"),
				err)
		}

		if err := tx.Set(
			client.Doc("folders/sports/folders/hockey"),
			&Folder{
				ID:       "hockey",
				ParentID: "sports",
			}); err != nil {
			return multierror.Append(
				errors.New("couldn't create hockey folder"),
				err)
		}

		if err := tx.Set(
			client.Doc("folders/sports/folders/baseball"),
			&Folder{
				ID:       "baseball",
				ParentID: "sports",
			}); err != nil {
			return multierror.Append(
				errors.New("couldn't create baseball folder"),
				err)
		}

		return nil
	}); err != nil {
		log.Fatalf("Failed to execute transaction: %v", err)
	}

	// Let's do a global query for all entities w/ NodeID = 1
	log.Info("Querying")

	// TODO Querying across subcollections is not currently supported in Cloud Firestore.
	//  If you need to query data across collections, use root-level collections
	iter := client.Collection("folders").Where("ID", "==", "sports").Documents(ctx)
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
