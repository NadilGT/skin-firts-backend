package apiHandlers

// firebase.go — Firebase is no longer used for authentication.
// This file is kept as a stub so any remaining import references compile
// cleanly during the transition period.
//
// The InitFirebaseApp function now returns nil, nil indicating no Firebase
// dependency.  All Firebase token verification has been replaced by the
// local JWT middleware in the /auth package.

// InitFirebaseApp is a no-op stub. Returns nil to signal no Firebase app.
func InitFirebaseApp() (interface{}, error) {
	return nil, nil
}
