package users

import (
	"bytes"
	"fmt"
)

type User struct {
	Username string
	Uid      uint64

	PasswordHash []byte
	tokens       [][]byte

	//sm sync.Mutex
	//ss *net.Conn

	Msgs MsgStore

	DHPubKey []byte
}

// AddToken todo: optimize token addition and validation if necessary (possibly binary search)
//adds token passed to it to the user and returns the same token
func (u *User) AddToken(token []byte) []byte {
	u.tokens = append(u.tokens, token)
	return token
}

func (u *User) RemoveToken(token []byte) {
	for i, _ := range u.tokens {
		if bytes.Equal(u.tokens[i], token) {
			u.tokens[i] = u.tokens[len(u.tokens)-1]
			u.tokens = u.tokens[:len(u.tokens)-1]
		}
	}
}

func (u *User) ValidToken(token []byte) bool {
	for _, t := range u.tokens {
		if bytes.Equal(t, token) {
			fmt.Println("valid token")
			return true
		}
	}
	return false
}
