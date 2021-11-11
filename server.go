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
			case protocol.CmdRegCheckUsernameAvailability:
				if err == nil {
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusUsernameTaken})
				} else {
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusUsernameOK})

				}
				break
			case protocol.CmdRegRegister:
				if err != nil { //user doesn't exist; available to register
					if len(req.S.Pubkey) != protocol.KeyLength { //key needs to exist and be the right length for curve25519
						protocol.SResp(c, &pb.SResp{Status: protocol.ErrInvalidKeyLength})
						continue
					}
					cusr.Username = req.S.Username
					cusr.Uid = crypto.RandUint64()
					cusr.PasswordHash = crypto.PWHash(cusr.Uid, req.S.Password)
					cusr.Msgs = users.NewListStore()
					cusr.DHPubKey = req.S.Pubkey
					t := cusr.AddToken(crypto.RandToken(protocol.TokenLength))
					s.AddUser(cusr)
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK, Token: t})
					goto regFin
				} else { //user can't register with an existing username
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusUsernameTaken})
				}
				break
			case protocol.CmdRegLogin:
				if err == nil { //user exists
					if cusr.ValidToken(req.Token) {
						protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK, Token: req.Token}) //token login
						goto regFin
					} else if bytes.Equal(cusr.PasswordHash, crypto.PWHash(cusr.Uid, req.S.Password)) { //password login
						t := cusr.AddToken(crypto.RandToken(protocol.TokenLength))
						s.UpdateUser(cusr)
						protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK, Token: t})
						goto regFin
					} else { //no login
						protocol.SResp(c, &pb.SResp{Status: protocol.ErrIncorrectCredentials})
					}
				} else { //user doesn't exist
					protocol.SResp(c, &pb.SResp{Status: protocol.ErrNotFound})
				}
				break
			default: //only registration/login commands are available
				protocol.SResp(c, &pb.SResp{Status: protocol.ErrInvalidCommand})
			}
		} else { //prior to authenticating, user can't send messages
			protocol.SResp(c, &pb.SResp{Status: protocol.ErrUnauthorized})
		}
	}
regFin:
	log.Println("User", cusr.Username, "logged in from ", c.RemoteAddr())

	for {
		var u users.User
		req.Reset()
		err = conn.RecvPB(c, &req)
		if err != nil {
			fmt.Println("error getting packet from client:", err)
			continue
		} else if cusr.ValidToken(req.Token) {
			for _, dc := range req.Dmcs {
				if dc.Buffered {
					for _, dm := range dc.D {
						if u, err = s.GetUserByUID(dm.Recipient); err == nil {
							dm.Origin = cusr.Uid
							u.Msgs.QueueMsg(dm)
						} else {
							protocol.SResp(c, &pb.SResp{Status: protocol.ErrNotFound})
						}
					}
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK})
				} else {
					for _, dm := range dc.D {
						if u, err = s.GetUserByUID(dm.Recipient); err == nil {
							//todo: send directly
						} else {
							protocol.SResp(c, &pb.SResp{Status: protocol.ErrNotFound})
						}
					}
				}
			}
			if req.S != nil {
				switch req.S.Command {
				case protocol.CmdGetUid: //get uid from username
					u, err = s.GetUserByUsername(req.S.Username)
					if err == nil {
						protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK, Uid: u.Uid})
					} else {
						protocol.SResp(c, &pb.SResp{Status: protocol.ErrNotFound})
					}
					break
				case protocol.CmdGetUsername: //get username from uid
					if u, err = s.GetUserByUID(req.S.Uid); err == nil { //user exists
						protocol.SResp(c, &pb.SResp{Status: protocol.StatusOK, Username: u.Username})
					} else {
						protocol.SResp(c, &pb.SResp{Status: protocol.ErrNotFound})
					}
					break
				case protocol.CmdGetMsgs:
					t := 0
					m := make([]*pb.DM, 0)
					l := cusr.Msgs.Len()
					for i := 0; i < l; i++ {
						d := cusr.Msgs.DequeueMsg()
						m = append(m, d)
						if t += len(d.Data); t > protocol.MaxResponseLength {
							break
						}
					}
					//a := protocol.MaxTries
					//for ; a > 0; a-- {
					//	conn.SendPB(c, &pb.RMes{R: &pb.SResp{Status: protocol.StatusDM, Dms: m, Remaining: uint32(cusr.Msgs.Len())}})
					//	conn.RecvPB(c, &req)
					//	if int(req.S.Addl) == len(m) {
					//		protocol.SResp(c, &pb.SResp{Status: protocol.StatusDMGood})
					//		break
					//	}
					//	fmt.Println("error sending dms to", c.RemoteAddr())
					//}
					//if a == 0 {
					//	fmt.Println("gave up sending messages to", cusr.Username, "at", c.RemoteAddr())
					//}
					conn.SendPB(c, &pb.RMes{R: &pb.SResp{Status: protocol.StatusDM, Dms: m, Remaining: uint32(cusr.Msgs.Len())}})
					break
				case protocol.CmdLogout:
					s.UpdateUser(cusr)
					protocol.SResp(c, &pb.SResp{Status: protocol.StatusGoodbye})
					return
				default:
					protocol.SResp(c, &pb.SResp{Status: protocol.ErrInvalidCommand})
				}
			}
		} else {
			protocol.SResp(c, &pb.SResp{Status: protocol.ErrUnauthorized})
		}
	}
}
