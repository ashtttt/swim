package swim

import (
	"errors"
	"fmt"
	"log"
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
		fmt.Println(err)
		return err
	}
	if msgType == pingMsgType {
		hb := new(ping)

		deserialize(msgBytes, hb)
		if err != nil {
			fmt.Println(err)
			return err
		}
		ack := ackMessage{SeqNo: hb.SeqNo, Name: hb.Name, PayLod: s.nodes}
		buff := serialize(&ack, ackMsgType)

		err = s.sendMessage(buff, conn)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	if msgType == ackMsgType {
		ack := new(ackMessage)
		err := deserialize(msgBytes, ack)

		if ack.SeqNo == s.seqNo {
			//err := s.setAlive(&Node{Name: ack.Name, Addr: conn.})
			err = s.handleAck(*ack)
			if err != nil {
				return err
			}
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

func (s *Swim) handleAck(ack ackMessage) error {

	for _, remoteNode := range ack.PayLod {
		status, i := s.isLocalNode(remoteNode)
		if status {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			s.nodes = append(s.nodes, remoteNode)
		} else {
			s.nodes = append(s.nodes, remoteNode)
		}
	}
	log.Print(s.nodes[0])
	return nil
}

func (s *Swim) setAlive(node *Node) error {
	staus, index := s.isLocalNode(node)
	if staus {
		localNode := s.nodes[index]
		localNode.Status = node.Status
	} else {
		s.nodes = append(s.nodes, node)
	}
	return nil
}
