package conn

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
)

func SendDelimited(con net.Conn, b []byte) error {
	var buf = make([]byte, binary.MaxVarintLen64)

	t := binary.PutUvarint(buf, uint64(len(b))) //store varint in buf
	buf = append(buf[:t], b...)                 //remove unused varint bytes before appending data

	_, err := con.Write(buf)
	return err
}

func ReadDelimited(con net.Conn) ([]byte, error) {
	var l uint64
	var res []byte

	//get len of rest of message
	v := bufio.NewReader(con)
	l, err := binary.ReadUvarint(v)

	//read rest of message
	res = make([]byte, l)
	_, err = io.ReadFull(v, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RecvPB(c net.Conn, req proto.Message) error {
	res, err := ReadDelimited(c)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(res, req)

	//piss
	g, _ := json.MarshalIndent(req, "", "  ")
	println("received:", string(g))

	return err
}

func SendPB(c net.Conn, resp proto.Message) error {
	//piss
	g, _ := json.MarshalIndent(resp, "", "  ")
	println("sent:", string(g))

	res, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	err = SendDelimited(c, res)
	return err
}
