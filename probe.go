package swim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"
)

func (s *Swim) initProbing() error {
	ticker := time.NewTicker(s.config.ProbingInterval)
	s.ticker = ticker
	for _ = range s.ticker.C {
		node := s.nextTarget()
		if node != nil {
			go s.probe(node)
		}
	}
	return nil
}

func (s *Swim) probe(node *Node) error {

	if len(node.Addr) <= 0 {
		return errors.New("Node address cant be empty")
	}
	ping := ping{SeqNo: s.nextSeqNo()}

	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, &ping)

	conn, err := s.transport.getDailer(node, s.config.ProbeTimeout)
	if err != nil {
		// TODO: Log Error
		return err
	}
	s.sendMessage(buff.Bytes(), conn)
	if err != nil {
		//TODO :  log error
		return err
	}
	return nil
}

func (s *Swim) sendMessage(data []byte, conn net.Conn) error {
	defer conn.Close()
	conn.Write(data)
	return nil
}
