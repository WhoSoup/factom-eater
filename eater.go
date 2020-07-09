package eater

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/WhoSoup/factom-eater/eventmessages"
)

// ProtocolVersion is the highest version of the Live Event API understood by the eater
const ProtocolVersion byte = 1

// Eater is an endpoint that provides boilerplate functionality for reading events from the Live Event API
// and making it accessible via a golang channel.
type Eater struct {
	listener net.Listener
	stopper  chan bool
	once     sync.Once

	events chan *eventmessages.FactomEvent
}

// Launch a new Eater.
// The host field should correspond to the `EventReceiverHost` and `EventReceiverPort` settings in factomd.conf.
// Returns an error if unable to bind to the socket.
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

// Stop the eater. Will tear down the listener and close the receiver channel.
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

// listen to a socket for inbound connections.
func (e *Eater) listen() {
	for {
		con, err := e.listener.Accept()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				time.Sleep(time.Millisecond * 250)
				continue
			}
			return
		}

		go e.consume(con)
		defer con.Close()
	}
}

// decodes live feed signals and sends them to channel.
func (e *Eater) consume(con net.Conn) {
	metabuf := make([]byte, 5) // 1 byte version + 4 byte little endian uint32
	for {
		select {
		case <-e.stopper:
			return
		default:
		}

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

// Reader returns the read-only channel that all received events are sent to.
// All calls to Reader() return the same channel.
func (e *Eater) Reader() <-chan *eventmessages.FactomEvent {
	return e.events
}
