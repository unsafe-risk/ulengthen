package hashalg

import (
	"testing"
)

func BenchmarkArgon2ID_V1(b *testing.B) {
	h := Argon2id_v1{}
	pw := []byte("password")
	salt := []byte("salt_value")
	for i := 0; i < b.N; i++ {
		h.Hash(pw, salt)
	}
}
