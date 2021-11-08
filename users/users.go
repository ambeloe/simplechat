package users

import "fmt"

type User struct {
	Username string
	Uid      uint64

	PasswordHash []byte
	tokens       []string

	//sm sync.Mutex
	//ss *net.Conn

	Msgs MsgStore

	DHPubKey []byte
}

// AddToken todo: optimize token addition and validation if necessary (possibly binary search)
//adds token passed to it to the user and returns the same token
func (u *User) AddToken(token string) string {
	u.tokens = append(u.tokens, token)
	return token
}

func (u *User) RemoveToken(token string) {
	for i, _ := range u.tokens {
		if u.tokens[i] == token {
			u.tokens[i] = u.tokens[len(u.tokens)-1]
			u.tokens = u.tokens[:len(u.tokens)-1]
		}
	}
}

func (u *User) ValidToken(token string) bool {
	for _, t := range u.tokens {
		if t == token {
			fmt.Println("valid token")
			return true
		}
	}
	return false
}
