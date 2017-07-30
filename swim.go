package swim

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"
)

type Swim struct {
	config    *Config
	transport *Transport
	nodes     []*Node

	ticker *time.Ticker

	targetIndex int
	seqNo       int

	locker sync.Mutex
}

type Node struct {
	Name   string
	Addr   string
	Port   int
	Status string // alive, suspect, dead
}

func (n *Node) Address() string {
	return net.JoinHostPort(n.Addr, strconv.Itoa(int(n.Port)))
}

func Init(config *Config) (*Swim, error) {
	if config == nil {
		config = config.DefaultConfig()
	}
	swim := &Swim{
		config: config,
	}
	if swim.transport == nil {
		tconfig := &TransportConfig{
			BindAddr: config.BindAddr,
			BindPort: config.BindPort,
		}
		t, err := NewTransport(tconfig)
		if err != nil {
			return nil, err
		}
		swim.transport = t
	}
	go swim.processRequests()
	go swim.initProbing()
	return swim, nil
}

// Join
func (s *Swim) Join(knownHosts []string) error {
	if len(knownHosts) <= 0 {
		return errors.New("Atleast one known host is required")
	}

	for _, host := range knownHosts {
		node := &Node{
			Addr: host,
			Port: s.config.BindPort,
		}
		err := s.probe(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Swim) Members() []*Node {
	return s.nodes
}

func (s *Swim) processRequests() {
	for {
		select {
		case con := <-s.transport.tcpCh:
			go s.processMessages(con)
		}
	}
}

func (s *Swim) processMessages(conn net.Conn) error {
	conn.SetDeadline(time.Now().Add(s.config.ProbeTimeout))
	defer conn.Close()

	msgType, msgBytes, err := s.readMessage(conn)

	if err != nil {
		// TODO: Log error
		return err
	}
	switch msgType {
	case pingMsgType:
		err := s.processPing(conn, msgBytes)
		if err != nil {
			//TODO: log error
			return err
		}
	case ackMsgType:
		// TODO: handle ack
	}
	return nil
}

func (s *Swim) nextTarget() *Node {
	if len(s.nodes) > 0 {
		if s.targetIndex == 0 {
			s.targetIndex = len(s.nodes) - 1
		} else {
			s.targetIndex = s.targetIndex - 1
		}
		node := s.nodes[s.targetIndex]
		if node.Name == s.config.Name {
			return s.nextTarget()
		}
		return node
	}
	return nil
}

func (s *Swim) nextSeqNo() int {
	s.seqNo = s.seqNo + 1
	return s.seqNo
}

func (s *Swim) processPing(conn net.Conn, msg []byte) error {
	hb := new(ping)

	err := deserialize(msg, hb)
	if err != nil {
		// TODO: Log Error
		return err
	}
	ack := ackMessage{SeqNo: hb.SeqNo, Name: hb.Name, PayLod: s.nodes}
	buff := serialize(&ack, ackMsgType)

	err = s.sendMessage(buff, conn)
	if err != nil {
		// TODO: log error
		return err
	}
	return nil
}
