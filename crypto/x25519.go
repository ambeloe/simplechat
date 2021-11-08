package crypto

import (
	"golang.org/x/crypto/curve25519"
	"io"
)

type X25519 struct {
	PubKey  []byte
	privKey []byte
}

//rand is the reader where key material is read from; make sure it's good random!
func NewPair(rand io.Reader) (*X25519, error) {
	var err error
	var x = X25519{}

	x.privKey = make([]byte, curve25519.ScalarSize)
	if _, err = io.ReadFull(rand, x.privKey); err != nil {
		return nil, err
	}

	x.PubKey, err = curve25519.X25519(x.privKey, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	return &x, nil
}
