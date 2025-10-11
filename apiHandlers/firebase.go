package apiHandlers

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitFirebaseApp() (*firebase.App, error) {
    opt := option.WithCredentialsFile("./skin-firts-firebase-adminsdk-fbsvc-fa502744d8.json")
    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        return nil, fmt.Errorf("error initializing app: %v", err)
    }
    return app, nil
}
