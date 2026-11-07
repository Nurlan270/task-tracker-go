package hasher

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

type ArgonHasher struct {
	params *ArgonParams
	sem    chan struct{}
}

func NewArgonHasher(params *ArgonParams) *ArgonHasher {
	if params == nil {
		params = NewArgonParams(32*1024, 2, 2, 16, 32, 3)
	}

	return &ArgonHasher{
		params: params,
		sem:    make(chan struct{}, params.SemCap),
	}
}

type ArgonParams struct {
	//	Describes the core parameters used by the Argon2id algorithm
	core *argon2id.Params

	//	Semaphore capacity. The Default value is 3. It's used to prevent memory overflow
	SemCap int
}

func NewArgonParams(mem, iter uint32, parallel uint8, saltLen, keyLen uint32, semCap int) *ArgonParams {
	return &ArgonParams{
		core: &argon2id.Params{
			Memory:      mem,
			Iterations:  iter,
			Parallelism: parallel,
			SaltLength:  saltLen,
			KeyLength:   keyLen,
		},
		SemCap: semCap,
	}
}

func (h ArgonHasher) Hash(v string) (string, error) {
	h.sem <- struct{}{}
	defer func() { <-h.sem }()

	hash, err := argon2id.CreateHash(v, h.params.core)
	if err != nil {
		return "", fmt.Errorf("argon2id: failed to create hash: %w", err)
	}

	return hash, nil
}

func (h ArgonHasher) Verify(v, hash string) (bool, error) {
	h.sem <- struct{}{}
	defer func() { <-h.sem }()

	match, err := argon2id.ComparePasswordAndHash(v, hash)
	if err != nil {
		return false, fmt.Errorf("argon2id: failed to verify value against hash: %w", err)
	}

	return match, nil
}
