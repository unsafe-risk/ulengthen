package hashalg

import "golang.org/x/crypto/argon2"

type Argon2id_v1 struct{}

func (a Argon2id_v1) Hash(pw, salt []byte) []byte {
	return argon2.IDKey(pw, salt, 64, 1024, 1, 32)
}
