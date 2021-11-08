package protocol

import "irchat/protocol/pb"

func SimpleDM(token string, recipient uint64, data []byte) *pb.Mes {
	var m = &pb.Mes{}
	m.Token = token
	m.Dmcs = []*pb.DmCont{{Buffered: true, M: []*pb.Dm{{Recipient: recipient, Data: data}}}}
	return m
}
