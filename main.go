package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type User struct {
	ID string
}

func (u User) String() string {
	return fmt.Sprintf("user %s", u.ID)
}

type Session struct {
	ID string
}

func (u Session) String() string {
	return fmt.Sprintf("session %s", u.ID)
}

func main() {
	log.Info("hi")

	ctx := context.TODO()

	// Create client
	log.Info("Creating client")
	client, err := firestore.NewClient(ctx, "irisvr-shared")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Close client when done.
	defer client.Close()

	err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		userDocRef := client.Collection("users").Doc("1")
		if err := tx.Set(userDocRef, &User{
			ID: "1",
		}); err != nil {
			return err
		}

		sessionDocRef := userDocRef.Collection("sessions").Doc("1")
		if err := tx.Set(sessionDocRef, &Session{
			ID: "1",
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		userDocRef := client.Collection("users").Doc("1")

		ds, err := tx.Get(userDocRef)
		if err != nil {
			return err
		}
		var u User
		if err := ds.DataTo(&u); err != nil {
			return err
		}
		log.Infof("tx user = %s", u)

		clientIter := userDocRef.Collection("sessions").Documents(ctx)
		for {
			ds, err := clientIter.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				log.Warnf("client iter error: %v", err)
				break
			} else if ds == nil {
				log.Warn("nil doc snapshot")
				break
			}

			var s Session
			if err := ds.DataTo(&s); err != nil {
				return err
			}
			log.Infof("client session = %s", s)
		}

		log.Info("ITERATING WITH TRANSACTION...")

		txIter := tx.Documents(userDocRef.Collection("sessions"))
		for {
			ds, err := txIter.Next()
			if err == iterator.Done {
				break
			} else if err != nil {
				log.Warnf("tx iter error: %v", err)
				break
			} else if ds == nil {
				log.Warn("nil doc snapshot")
				break
			}

			var s Session
			if err := ds.DataTo(&s); err != nil {
				return err
			}
			log.Infof("tx session = %s", s)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}

