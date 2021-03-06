package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"golang.org/x/crypto/argon2"
	"io"
	"irchat/protocol"
)

func PWHash(uid uint64, pass []byte) []byte {
	salt := make([]byte, 8)
	binary.BigEndian.PutUint64(salt, uid)
	return argon2.IDKey(pass, salt, 10, 64*1024, 2, protocol.PWHashLen)
}

func RandUint64() uint64 {
	var uD = make([]byte, 8)
	_, err := io.ReadAtLeast(rand.Reader, uD, 8)
	if err != nil {
		panic("error reading randomness")
	}
	return binary.BigEndian.Uint64(uD)
}

func RandToken(len int) []byte {
	var tok = make([]byte, len)
	_, err := io.ReadAtLeast(rand.Reader, tok, len)
	if err != nil {
		panic("error reading randomness")
	}
	return tok
}
