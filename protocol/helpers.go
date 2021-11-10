package protocol

import (
	"irchat/conn"
	"irchat/protocol/pb"
	"net"
)

func CReq(c net.Conn, token []byte, r *pb.CReq) error {
	var cr = pb.Mes{Token: token, S: r}
	return conn.SendPB(c, &cr)
}

func CSendBuffDM(c net.Conn, token []byte, buff bool, recipient uint64, data []byte, group uint64) error {
	return conn.SendPB(c, &pb.Mes{Token: token, Dmcs: []*pb.DCont{{Buffered: buff, D: []*pb.DM{{Recipient: recipient, Data: data, GroupID: group}}}}})
}

func SResp(c net.Conn, s *pb.SResp) error {
	return conn.SendPB(c, &pb.RMes{R: s})
}
