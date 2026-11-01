package hasher

type Hasher interface {
	//	Hash returns hashed result of passed value
	Hash(v string) (string, error)

	//	Verify verifies passed value against hash
	//	and returns true if hashes match, false otherwise
	Verify(v, hash string) (bool, error)
}
