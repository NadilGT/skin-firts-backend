package dao

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestRawUserListing_DecodeCreatedAtFromBSONDateTime(t *testing.T) {
	expected := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	raw, err := bson.Marshal(bson.M{
		"userId":    "USR-001",
		"name":      "Test User",
		"email":     "test@example.com",
		"role":      "admin",
		"branchIds": []string{"BRN-001"},
		"status":    "ACTIVE",
		"createdAt": expected,
	})
	if err != nil {
		t.Fatalf("failed to marshal BSON fixture: %v", err)
	}

	var listing RawUserListing
	if err := bson.Unmarshal(raw, &listing); err != nil {
		t.Fatalf("failed to unmarshal BSON into RawUserListing: %v", err)
	}

	if listing.UserID != "USR-001" {
		t.Fatalf("unexpected userId: %s", listing.UserID)
	}
	if !listing.CreatedAt.Equal(expected) {
		t.Fatalf("unexpected createdAt: got %s want %s", listing.CreatedAt, expected)
	}
}

func TestRawUserListing_JSONMarshalsCreatedAtAsString(t *testing.T) {
	listing := RawUserListing{
		UserID:    "USR-001",
		Name:      "Test User",
		Email:     "test@example.com",
		Role:      "admin",
		BranchIds: []string{"BRN-001"},
		Status:    "ACTIVE",
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
	}

	data, err := json.Marshal(listing)
	if err != nil {
		t.Fatalf("failed to marshal RawUserListing to JSON: %v", err)
	}

	jsonText := string(data)
	if !strings.Contains(jsonText, `"createdAt":"2026-01-02T03:04:05Z"`) {
		t.Fatalf("createdAt was not serialized as an RFC3339 string: %s", jsonText)
	}
}
