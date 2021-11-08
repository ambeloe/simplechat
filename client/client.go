package main

import (
	"crypto/rand"
	"fmt"
	"github.com/ambeloe/cliui"
	"irchat/conn"
	"irchat/crypto"
	"irchat/protocol"
	"irchat/protocol/pb"
	"net"
)

type Usr struct {
	Username string

	Token string

	Key *crypto.X25519
}

var err error

var c net.Conn
var resp pb.ServResp
var u Usr

func main() {
	a := cliui.UI{Name: "chat"}

	a.Add("c", connect)
	a.Add("s", send)
	a.Add("l", login)
	a.Add("r", register)

	a.Run()
}

func connect(s []string) {
	if len(s) != 1 {
		fmt.Println("args: address:port (127.0.0.1:2500)")
		return
	}
	c, err = net.Dial("tcp", s[0])
	if err != nil {
		fmt.Println("error connecting to server:", err)
		return
	}
	fmt.Println("connected to", s[0])
}

func login(s []string) {
	if len(s) != 2 {
		fmt.Println("args: username password")
		return
	}
	conn.SendPB(c, &pb.Mes{S: &pb.ServReq{Command: protocol.CmdLogin, Username: s[0], Password: []byte(s[1])}})
	conn.RecvPB(c, &resp)
	if resp.Status != protocol.StatusOK {
		fmt.Println("error logging in:", protocol.ErrToString(resp.Status))
		return
	}
	u.Username = s[0]
	u.Token = resp.Token
}

func register(s []string) {
	if len(s) != 2 {
		fmt.Println("args: username password")
		return
	}
	u.Username = s[0]
	conn.SendPB(c, &pb.Mes{S: &pb.ServReq{Command: protocol.CmdCheckUsernameAvailability, Username: u.Username}})
	conn.RecvPB(c, &resp)
	if resp.Status == protocol.StatusUsernameTaken {
		fmt.Println("that usename is taken")
		return
	}
	u.Key, _ = crypto.NewPair(rand.Reader)
	conn.SendPB(c, &pb.Mes{S: &pb.ServReq{Command: protocol.CmdRegister, Username: u.Username, Password: []byte(s[1]), Pubkey: u.Key.PubKey}})
	conn.RecvPB(c, &resp)
	if resp.Status != protocol.StatusOK {
		fmt.Println("error registering:", protocol.ErrToString(resp.Status))
		return
	}
	u.Token = resp.Token
	fmt.Println("successfully registered", u.Username)
}

func send(s []string) {
	if len(s) != 2 {
		fmt.Println("args: recipient message")
		return
	}
	conn.SendPB(c, &pb.Mes{S: &pb.ServReq{Command: protocol.CmdGetUsername, Username: s[0]}})
	if resp.Status != protocol.StatusOK {
		fmt.Println("error getting user uid:", protocol.ErrToString(resp.Status))
		return
	}
	conn.SendPB(c, protocol.SimpleDM(u.Token, resp.Uid, []byte(s[1])))
	conn.RecvPB(c, &resp)
	if resp.Status != protocol.StatusOK {
		fmt.Println("error sending message:", protocol.ErrToString(resp.Status))
		return
	}
}
