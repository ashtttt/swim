package swim

import (
	"log"
	"net"
	"strconv"
	"time"
)

type Swim struct {
	config    *Config
	transport *Transport
	Nodes     []Node

	ticker *time.Ticker

	targetIndex int
	seqNo       int
}

type Node struct {
	Name string
	Addr string
	Port int
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

	return nil, nil
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
	buf := make([]byte, 1024)

	reqLen, err := conn.Read(buf)
	log.Print(reqLen)
	if err != nil {
		// TODO:log erro here
		log.Print(err)
		return err
	}
	conn.Write([]byte("Message received."))
	return nil
}

func (s *Swim) nextTarget() *Node {
	if len(s.Nodes) > 0 {
		if s.targetIndex == 0 {
			s.targetIndex = len(s.Nodes) - 1
		} else {
			s.targetIndex = s.targetIndex - 1
		}
		return &s.Nodes[s.targetIndex]
	}
	return nil
}

func (s *Swim) nextSeqNo() int {
	return s.seqNo + 1
}
