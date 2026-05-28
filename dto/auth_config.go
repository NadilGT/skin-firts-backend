package dto

// AuthConfig holds the configuration needed for JWT-based auth.
// FirebaseProjectID is kept (but unused) so any code referencing it still compiles
// during the transition period.
type AuthConfig struct {
	JWTSecret         string `json:"jwtSecret"`
	FirebaseProjectID string `json:"firebaseProjectId,omitempty"` // deprecated — kept for compat
}
