package swim

import (
	"log"
	"net"
)

type Swim struct {
	config    *Config
	transport *Transport
	logger    *log.Logger
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
			logger:   swim.logger,
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
