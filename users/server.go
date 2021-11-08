package users

import (
	"encoding/json"
	"errors"
	"irchat/crypto"
	"os"
	"sync"
)

type Server struct {
	ServerUID    uint64
	ServerName   string
	ServerDomain string

	usersByUID      sync.Map
	usersByUsername sync.Map

	//delUsers uint32 //sort of a fragmentation counter
	Users []User
}

func InitServer(name, domain string) *Server {
	var s = Server{
		ServerUID:       crypto.RandUint64(),
		ServerDomain:    domain,
		usersByUID:      sync.Map{},
		usersByUsername: sync.Map{},
		//delUsers:        0,
		Users: make([]User, 0),
	}
	return &s
}

//todo: will currently shit pant
func (s *Server) StoreServer(fn string) error {
	res, err := json.Marshal(s)
	if err != nil {
		return err
	}
	err = os.WriteFile(fn, res, 0644)
	return err
}

func (s *Server) LoadServer(fn string) error {
	in, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, s)
	return err
}

func (s *Server) GetUserByUID(uid uint64) (User, error) {
	if i, is := s.usersByUID.Load(uid); is {
		return s.Users[i.(int)], nil
	}
	return User{}, errors.New("user not found")
}

func (s *Server) GetUserByUsername(name string) (User, error) {
	if i, is := s.usersByUsername.Load(name); is {
		return s.Users[i.(int)], nil
	}
	return User{}, errors.New("user not found")
}

// AddUser not really needed anymore; keeping anyway
func (s *Server) AddUser(usr User) {
	s.Users = append(s.Users, usr)
	s.usersByUID.Store(usr.Uid, len(s.Users)-1)
	s.usersByUsername.Store(usr.Username, len(s.Users)-1)
}

func (s *Server) UpdateUser(usr User) {
	i, is := s.usersByUID.Load(usr.Uid)
	if is {
		s.Users = append(s.Users, usr)
		s.usersByUID.Store(usr.Uid, len(s.Users)-1)
		s.usersByUsername.Store(usr.Username, len(s.Users)-1)
	} else {
		s.Users[i.(int)] = usr
	}
}

//unused for now
//func (s *Server) DeleteUser(uid uint64) error {
//	if i, is := s.usersByUID.Load(uid); is {
//		s.usersByUID.Delete(uid)
//		s.usersByUsername.Delete(s.Users[i.(int)].Username)
//		atomic.AddUint32(&s.delUsers, 1)
//		return nil
//	}
//	return errors.New("user not found")
//}
