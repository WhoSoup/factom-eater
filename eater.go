package eater

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/WhoSoup/factom-eater/eventmessages"
)

const ProtocolVersion byte = 1

type Eater struct {
	listener net.Listener
	stopper  chan bool
	once     sync.Once

	events chan *eventmessages.FactomEvent
}

func Launch(host string) (*Eater, error) {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}

	e := new(Eater)
	e.stopper = make(chan bool)
	e.events = make(chan *eventmessages.FactomEvent)
	e.listener = listener
	go e.listen()
	return e, nil
}

func (e *Eater) Stop() error {
	e.once.Do(func() {
		close(e.stopper)
		close(e.events)
	})

	if err := e.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (e *Eater) listen() {
	for {
		con, err := e.listener.Accept()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				continue
			}
			return
		}

		go e.consume(con)
	}
}

func (e *Eater) consume(con net.Conn) {

	metabuf := make([]byte, 5) // 1 byte version + 4 byte little endian uint32
	for {
		if _, err := io.ReadFull(con, metabuf); err != nil {
			return
		}

		if metabuf[0] != ProtocolVersion {
			return
		}

		rawMsg := make([]byte, binary.LittleEndian.Uint32(metabuf[1:]))
		if _, err := io.ReadFull(con, rawMsg); err != nil {
			return
		}

		msg := new(eventmessages.FactomEvent)
		if err := msg.Unmarshal(rawMsg); err != nil {
			continue
		}

		e.events <- msg
	}
}

func (e *Eater) Reader() <-chan *eventmessages.FactomEvent {
	return e.events
}
