package token

import (
	"testing"
	"time"
)

func TestManagerGenerateAndValidateAccessToken(t *testing.T) {
	manager, err := NewManager("", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	pair, err := manager.GenerateTokenPair("user_1", "v1")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	claims, err := manager.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != "user_1" || claims.Version != "v1" || claims.Type != TypeAccess {
		t.Fatalf("ValidateAccessToken() claims = %+v, want user/version/access", claims)
	}
}

func TestManagerRejectsInvalidAccessToken(t *testing.T) {
	manager, err := NewManager("", time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if _, err := manager.ValidateAccessToken("bad-token"); err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error")
	}
}
