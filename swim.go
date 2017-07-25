package swim

import (
	"errors"
	"net"
	"strconv"
	"time"
)

type Swim struct {
	config    *Config
	transport *Transport
	nodes     []*Node

	ticker *time.Ticker

	targetIndex int
	seqNo       int
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
	defer conn.Close()
	messageBytes, err := s.readMessage(conn)
	if err != nil {
		return err
	}
	pingType, msgType := convertToObj(messageBytes)

	if msgType == pingMsgType {
		pingMsg := pingType.(ping)
		ack := ackMessage{SeqNo: pingMsg.SeqNo, PayLod: s.nodes}
		buff := converToBytes(&ack, ackMsgType)
		_, err = s.sendMessage(buff, conn)
		if err != nil {
			return err
		}
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
