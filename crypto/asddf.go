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

func RandString(len int) string {
	var str = make([]byte, len)
	_, err := io.ReadAtLeast(rand.Reader, str, len)
	if err != nil {
		panic("error reading randomness")
	}
	//possibly optimize with stringbuilder if need be
	return string(str)
}

func RandToken(len int) string {
	var str = make([]byte, len)
	_, err := io.ReadAtLeast(rand.Reader, str, len)
	if err != nil {
		panic("error reading randomness")
	}
	for i, _ := range str {
		str[i] = str[i] >> 1
	}
	//possibly optimize with stringbuilder if need be
	return string(str)
}
