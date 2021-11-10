package users

import (
	"container/list"
	"encoding/json"
	"irchat/protocol/pb"
)

type MsgStore interface {
	json.Marshaler
	json.Unmarshaler

	QueueMsg(cont *pb.DM)
	DequeueMsg() *pb.DM
	Len() int
}

type ListStore struct {
	l *list.List
}

func NewListStore() MsgStore {
	var m = ListStore{l: list.New()}
	return m
}

func (m ListStore) MarshalJSON() ([]byte, error) {
	if m.l.Len() == 0 {
		return []byte("[]"), nil
	}
	arr := make([]pb.DM, m.l.Len())
	i := 0
	e := m.l.Front()
	for e != nil {
		arr[i] = e.Value.(pb.DM)
		i++
		e = e.Next()
	}
	r, err := json.Marshal(arr)
	return r, err
}

func (m ListStore) UnmarshalJSON(b []byte) error {
	m.l = list.New()
	var arr []pb.DM
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return err
	}
	for _, e := range arr {
		m.l.PushBack(e)
	}
	return nil
}

func (m ListStore) QueueMsg(p *pb.DM) {
	m.l.PushBack(*p)
}

func (m ListStore) DequeueMsg() *pb.DM {
	return m.l.Remove(m.l.Front()).(*pb.DM)
}

func (m ListStore) Len() int {
	return m.l.Len()
}
