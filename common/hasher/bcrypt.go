package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

const cost = bcrypt.DefaultCost + 2

func (b BcryptHasher) Hash(v string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(v), cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: failed to create hash: %w", err)
	}

	return string(hash), nil
}

func (b BcryptHasher) Verify(v, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(v)); err != nil {
		return fmt.Errorf("bcrypt: failed to verify hash: %w", err)
	}

	return nil
}
