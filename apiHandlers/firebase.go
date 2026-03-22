package apiHandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitFirebaseApp() (*firebase.App, error) {
	// Load JSON from env
	creds := os.Getenv("FIREBASE_CREDENTIALS")
	if creds == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS not found in environment")
	}

	storageBucket := os.Getenv("FIREBASE_STORAGE_BUCKET")
	if storageBucket == "" {
		fmt.Println("WARNING: FIREBASE_STORAGE_BUCKET not found in environment. Image uploading will fail.")
	}

	// Validate JSON
	var tmp map[string]interface{}
	if err := json.Unmarshal([]byte(creds), &tmp); err != nil {
		return nil, fmt.Errorf("invalid firebase json: %v", err)
	}

	// Create app
	config := &firebase.Config{
		StorageBucket: storageBucket,
	}
	opt := option.WithCredentialsJSON([]byte(creds))
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase init failure: %v", err)
	}

	return app, nil
}
