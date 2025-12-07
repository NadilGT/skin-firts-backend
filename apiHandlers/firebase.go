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

	// Validate JSON
	var tmp map[string]interface{}
	if err := json.Unmarshal([]byte(creds), &tmp); err != nil {
		return nil, fmt.Errorf("invalid firebase json: %v", err)
	}

	// Create app
	opt := option.WithCredentialsJSON([]byte(creds))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase init failure: %v", err)
	}

	return app, nil
}
