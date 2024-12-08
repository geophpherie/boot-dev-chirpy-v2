package auth

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	passwords := []string{
		"as;ldfja0w89efjawj",
		"",
		"1",
	}
	for _, password := range passwords {
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatal("Couldn't hash password")
		}

		if err := CheckPasswordHash(password, hash); err != nil {
			t.Fatal("Hash mismatch!")
		}
	}
}

func TestJWTCreate(t *testing.T) {
	expectedUser, err := uuid.NewUUID()
	if err != nil {
		t.Error("can't make uuid")
	}

	expires := time.Duration(1) * time.Minute
	token, err := MakeJWT(expectedUser, "mySecretToken", expires)
	if err != nil {
		t.Errorf("error making jwt %v", err)
	}

	user, err := ValidateJWT(token, "mySecretToken")
	if err != nil {
		t.Errorf("error validating jwt %v", err)
	}

	if user != expectedUser {
		t.Fail()
	}
}

func TestJWTCreateFail(t *testing.T) {
	expectedUser, err := uuid.NewUUID()
	if err != nil {
		t.Error("can't make uuid")
	}

	expires := time.Duration(1) * time.Minute
	token, err := MakeJWT(expectedUser, "mySecretToken", expires)
	if err != nil {
		t.Errorf("error making jwt %v", err)
	}

	_, err = ValidateJWT(token, "mySecetToken")
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("token should be invalid")
	}
}

func TestBearerTokenParsing(t *testing.T) {
	header := http.Header{}

	header.Add("Authorization", "Bearer MYTOKENSTRING")

	token, err := GetBearerToken(header)
	if err != nil {
		t.Error("token parsing failed")
	}
	if token != "MYTOKENSTRING" {
		t.Fail()
	}
}

func TestNoBearerToken(t *testing.T) {
	header := http.Header{}

	header.Add("Authorization", "Bearer ")

	_, err := GetBearerToken(header)

	if !errors.Is(err, ErrNoBearerToken) {
		t.Error("parsing should have errored")
	}
}
