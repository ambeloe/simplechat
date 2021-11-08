package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"irchat/conn"
	"irchat/crypto"
	"irchat/protocol"
	"irchat/protocol/pb"
	"irchat/users"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	var p = flag.Uint("p", 2500, "port to listen on")
	var c = flag.String("c", "cert.pem", "certificate to use for client connections")
	var k = flag.String("k", "key.pem", "key to use for client connections")

	os.Exit(serve(*p, *c, *k))
}

func serve(port uint, certFile, keyFile string) int {
	l, err := net.Listen("tcp", "0.0.0.0:"+strconv.FormatUint(uint64(port), 10))
	if err != nil {
		fmt.Println("error listening on port", strconv.FormatUint(uint64(port), 10))
	}

	//init server
	var serv = users.InitServer("test server", "test.poggies.net")

	c, err := tls.LoadX509KeyPair(certFile, keyFile)

	var con net.Conn
	for {
		con, err = l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error accepting connection from", con.RemoteAddr())
			return 1
		}

		go (clientHandler)(con, &c, serv)
	}
}

func clientHandler(c net.Conn, cert *tls.Certificate, s *users.Server) {
	defer c.Close()

	//var cleanSlate = pb.Mes{}

	var err error
	var req = pb.Mes{}
	var cusr = users.User{}

	log.Println(c.RemoteAddr(), "connected")
	for { //unauthenticated
		req.Reset()
		conn.RecvPB(c, &req)
		if req.S != nil {
			cusr, err = s.GetUserByUsername(req.S.Username)
			switch req.S.Command {
			case protocol.CmdCheckUsernameAvailability:
				if err == nil {
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusUsernameTaken})
				} else {
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusUsernameOK})

				}
				break
			case protocol.CmdRegister:
				if err != nil { //user doesn't exist; available to register
					if len(req.S.Pubkey) != protocol.KeyLength { //key needs to exist and be the right length for curve25519
						conn.SendPB(c, &pb.ServResp{Status: protocol.ErrInvalidKeyLength})
						continue
					}
					cusr.Username = req.S.Username
					cusr.Uid = crypto.RandUint64()
					cusr.PasswordHash = crypto.PWHash(cusr.Uid, req.S.Password)
					cusr.Msgs = users.NewListStore()
					cusr.DHPubKey = req.S.Pubkey
					t := cusr.AddToken(crypto.RandToken(protocol.TokenLength))
					s.AddUser(cusr)
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK, Token: t})
					goto regFin
				} else { //user can't register with an existing username
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusUsernameTaken})
				}
				break
			case protocol.CmdLogin:
				if err == nil { //user exists
					if cusr.ValidToken(req.Token) {
						conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK, Token: req.Token}) //token login
						goto regFin
					} else if bytes.Equal(cusr.PasswordHash, crypto.PWHash(cusr.Uid, req.S.Password)) { //password login
						t := cusr.AddToken(crypto.RandString(protocol.TokenLength))
						s.UpdateUser(cusr)
						conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK, Token: t})
						goto regFin
					} else { //no login
						conn.SendPB(c, &pb.ServResp{Status: protocol.ErrIncorrectCredentials})
					}
				} else { //user doesn't exist
					conn.SendPB(c, &pb.ServResp{Status: protocol.ErrNotFound})
				}
				break
			default: //only registration/login commands are available
				conn.SendPB(c, &pb.ServResp{Status: protocol.ErrInvalidCommand})
			}
		} else { //prior to authenticating, user can't send messages
			conn.SendPB(c, &pb.ServResp{Status: protocol.ErrUnauthorized})
		}
	}
regFin:
	log.Println("User", cusr.Username, "logged in from ", c.RemoteAddr())

	for {
		var u users.User
		req.Reset()
		err = conn.RecvPB(c, &req)
		if err != nil {
			fmt.Println("shat pant:", err)
			continue
		} else if cusr.ValidToken(req.Token) {
			for _, dc := range req.Dmcs {
				if dc.Buffered {
					for _, dm := range dc.M {
						if u, err = s.GetUserByUID(dm.Recipient); err == nil {
							dm.Origin = cusr.Uid
							u.Msgs.QueueMsg(dm)
						} else {
							conn.SendPB(c, &pb.ServResp{Status: protocol.ErrNotFound})
						}
					}
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK})
				} else {
					for _, dm := range dc.M {
						if u, err = s.GetUserByUID(dm.Recipient); err == nil {
							//todo: send directly
						} else {
							conn.SendPB(c, &pb.ServResp{Status: protocol.ErrNotFound})
						}
					}
				}
			}
			if req.S != nil {
				switch req.S.Command {
				case protocol.CmdGetUid: //get uid from username
					if u, err = s.GetUserByUID(req.S.Uid); err == nil {
						conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK, Uid: u.Uid})
					} else {
						conn.SendPB(c, &pb.ServResp{Status: protocol.ErrNotFound})
					}
					break
				case protocol.CmdGetUsername: //get username from uid
					if u, err = s.GetUserByUID(req.S.Uid); err == nil { //user exists
						conn.SendPB(c, &pb.ServResp{Status: protocol.StatusOK, Username: u.Username})
					} else {
						conn.SendPB(c, &pb.ServResp{Status: protocol.ErrNotFound})
					}
					break
				case protocol.CmdLogout:
					conn.SendPB(c, &pb.ServResp{Status: protocol.StatusGoodbye})
					return
				default:
					conn.SendPB(c, &pb.ServResp{Status: protocol.ErrInvalidCommand})
				}
			}
		} else {
			conn.SendPB(c, &pb.ServResp{Status: protocol.ErrUnauthorized})
		}
	}
}
