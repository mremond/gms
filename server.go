package gms

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

// Magic numbers for GameMaker Studio 2
const (
	gmInit         = "GM:Studio-Connect"
	gmHeaderLen    = 12
	gmMagicNumber1 = uint32(0xdeadc0de)
	gmMagicNumber2 = uint32(0xcafebabe)
	gmMagicNumber3 = uint32(0xdeafbead)
	gmMagicNumber4 = uint32(0xf00dbeeb)
	gmMagicNumber5 = uint32(0x0000000c)
)

type Server struct {
	Raw       bool
	Port      int
	Clients   []Client
	nextIndex int
}

func (s Server) Start(port int, cb func(Message) error) {
	listenTo := fmt.Sprintf(":%d", port)
	log.Printf("Launching server on port %d", port)
	listener, err := net.Listen("tcp", listenTo)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		c := Client{conn: conn, Id: NewClientId(), raw: s.Raw, Index: s.nextIndex}
		s.nextIndex++
		go handleClient(c, s.Raw, cb)
	}
}

// Callback can be call concurrently as it is triggered by each client Go routine.
func handleClient(c Client, raw bool, cb func(Message) error) {
	defer c.conn.Close()
	if !raw {
		if err := c.handshake(); err != nil {
			// TODO: Only in debug mode
			log.Println("handshake failed:", err)
			return
		}
	}

	if err := cb(Message{Client: c, EventType: ClientConnect}); err != nil {
		cb(Message{Client: c, EventType: ClientDisconnect})
		return
	}

	reader := getReader(raw)
	for {
		if err := reader(c, cb); err != nil {
			// TODO: Only in debug mode
			log.Println(err)
			break
		}
	}

	cb(Message{Client: c, EventType: ClientDisconnect})
}

// getReader returns the appropriate data reader function, depending on
// the GM mode we want to support (GMProtocol or Raw)
func getReader(raw bool) func(Client, func(Message) error) error {
	if !raw {
		return readGMPacket
	}
	return readRawStream
}

// GM Protocol define packet size, so we can read the full packet in one go
func readGMPacket(c Client, cb func(Message) error) error {
	header := make([]byte, 12)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return err
	}
	header1 := uint32(binary.LittleEndian.Uint32(header[0:4]))
	header2 := uint32(binary.LittleEndian.Uint32(header[4:8]))
	header3 := uint32(binary.LittleEndian.Uint32(header[8:12]))
	if header1 != gmMagicNumber1 {
		return fmt.Errorf("packet contains invalid identifier/magic number")
	}
	if header2 != uint32(12) {
		return fmt.Errorf("packet header size is not 12")
	}
	payload := make([]byte, header3)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return err
	}
	if len(payload) != 0 {
		buffer := Packet{payload: payload[1:]}
		return cb(Message{Client: c, EventType: ClientData, DataType: int(payload[0]), Buffer: &buffer})
	}
	return nil
}

func readRawStream(c Client, cb func(Message) error) error {
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(c.conn, msgType); err != nil {
		return err
	}
	buffer := Stream{buffer: c.conn}
	return cb(Message{Client: c, EventType: ClientData, DataType: int(msgType[0]), Buffer: &buffer})
}

type Reader interface {
	Read(byteCount int) ([]byte, error)
}

type EventType int

const (
	ClientConnect EventType = iota
	ClientDisconnect
	ClientData
)

type Message struct {
	Client    Client
	EventType EventType
	DataType  int
	Buffer    Reader
}

//=============================================================================
// GM Protocol

type Packet struct {
	payload []byte
	// Message store read position to act as a reader
	readPos int
}

func (p *Packet) Read(byteCount int) ([]byte, error) {
	startPos := p.readPos
	nextPos := startPos + byteCount
	if nextPos > len(p.payload) {
		return []byte{}, io.EOF
	}
	p.readPos = nextPos
	return p.payload[startPos:nextPos], nil
}

//=============================================================================
// Raw protocol

type Stream struct {
	buffer io.Reader
}

func (s *Stream) Read(byteCount int) ([]byte, error) {
	data := make([]byte, byteCount)
	if _, err := io.ReadFull(s.buffer, data); err != nil {
		return []byte{}, err
	}
	return data, nil
}
