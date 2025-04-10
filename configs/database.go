package configs

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func NewFirestoreClient() *firestore.Client {
	ctx := context.Background()
	projectID := os.Getenv("FIREBASE_PROJECT_ID")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	firestoreClient := client
	log.Println("Firestore initialized")
	return firestoreClient
}
