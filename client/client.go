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

	Token []byte

	Key *crypto.X25519
}

var err error

var c net.Conn
var resp pb.RMes
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
	protocol.CReq(c, nil, &pb.CReq{Command: protocol.CmdRegLogin, Username: s[0], Password: []byte(s[1])})
	conn.RecvPB(c, &resp)
	if resp.R.Status != protocol.StatusOK {
		fmt.Println("error logging in:", protocol.ErrToString(resp.R.Status))
		return
	}
	u.Username = s[0]
	u.Token = resp.R.Token
}

func register(s []string) {
	if len(s) != 2 {
		fmt.Println("args: username password")
		return
	}
	u.Username = s[0]
	protocol.CReq(c, nil, &pb.CReq{Command: protocol.CmdRegCheckUsernameAvailability, Username: u.Username})
	conn.RecvPB(c, &resp)
	if resp.R.Status == protocol.StatusUsernameTaken {
		fmt.Println("that usename is taken")
		return
	}
	u.Key, _ = crypto.NewPair(rand.Reader)
	protocol.CReq(c, nil, &pb.CReq{Command: protocol.CmdRegRegister, Username: u.Username, Password: []byte(s[1]), Pubkey: u.Key.PubKey})
	conn.RecvPB(c, &resp)
	if resp.R.Status != protocol.StatusOK {
		fmt.Println("error registering:", protocol.ErrToString(resp.R.Status))
		return
	}
	u.Token = resp.R.Token
	fmt.Println("successfully registered", u.Username)
}

func send(s []string) {
	if len(s) != 2 {
		fmt.Println("args: recipient message")
		return
	}
	protocol.CReq(c, u.Token, &pb.CReq{Command: protocol.CmdGetUid, Username: s[0]})
	conn.RecvPB(c, &resp)
	if resp.R.Status != protocol.StatusOK {
		fmt.Println("error getting user uid:", protocol.ErrToString(resp.R.Status))
		return
	}
	protocol.CSendBuffDM(c, u.Token, true, resp.R.Uid, []byte(s[1]), 0)
	conn.RecvPB(c, &resp)
	if resp.R.Status != protocol.StatusOK {
		fmt.Println("error sending message:", protocol.ErrToString(resp.R.Status))
		return
	}
}
