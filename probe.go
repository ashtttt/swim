package swim

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const (
	pingMsgType uint8 = 0
	ackMsgType  uint8 = 1
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

	buff := converToBytes(&ping, pingMsgType)
	conn, err := s.transport.getDailer(node, s.config.ProbeTimeout)

	if err != nil {
		// TODO: Log Error
		return err
	}
	defer conn.Close()
	response, err := s.sendMessage(buff, conn)

	if err != nil {
		return err
	}
	ackType, msgType := convertToObj(response)
	log.Print(msgType)

	if msgType == ackMsgType {
		ack := ackType.(ackMessage)
		if ack.SeqNo == s.seqNo {
			err := s.setAlive(node)
			err = s.handleAck(ack)
			if err != nil {
				return err
			}
			log.Printf("Received ACK for sequence number: %d", s.seqNo)
		} else {
			return errors.New("Un-expected sequence number")
		}
		if err != nil {
			//TODO :  log error
			return err
		}
	}
	return nil
}

func (s *Swim) sendMessage(message []byte, conn net.Conn) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(s.config.ProbeTimeout))
	_, err := conn.Write(message)
	if err != nil {
		return nil, err
	}
	messageBytes, err := s.readMessage(conn)
	if err != nil {
		return nil, err
	}
	return messageBytes, nil
}

func (s *Swim) readMessage(conn net.Conn) ([]byte, error) {

	reader := bufio.NewReader(conn)
	buff := bytes.NewBuffer(nil)
	_, err := io.Copy(buff, reader)
	if err != nil {
		return nil, err
	}
	log.Print(buff.Bytes())
	return buff.Bytes(), nil

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
