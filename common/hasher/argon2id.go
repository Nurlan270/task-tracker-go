package hasher

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

type ArgonHasher struct{}

func NewArgonHasher() *ArgonHasher {
	return &ArgonHasher{}
}

// Default argon2id hashing parameters.
var params = &argon2id.Params{
	Memory:      32 * 1024, // 32 MiB
	Iterations:  2,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

//	Semaphore capacity = 2
//
// Max. 2 concurrent requests can hash password at the same time
// which will lead in 64 MiB memory usage; It's used to avoid memory overflow.
var sem = make(chan struct{}, 2)

func (h ArgonHasher) Hash(v string) (string, error) {
	sem <- struct{}{}
	defer func() { <-sem }()

	hash, err := argon2id.CreateHash(v, params)
	if err != nil {
		return "", fmt.Errorf("argon2id: failed to create hash: %w", err)
	}

	return hash, nil
}

func (h ArgonHasher) Verify(v, hash string) error {
	sem <- struct{}{}
	defer func() { <-sem }()

	if _, err := argon2id.ComparePasswordAndHash(v, hash); err != nil {
		return fmt.Errorf("argon2id: failed to verify value against hash: %w", err)
	}

	return nil
}
