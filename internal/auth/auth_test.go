package auth

import "testing"

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
