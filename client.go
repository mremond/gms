package gms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/satori/go.uuid"
)

type UUID [16]byte

// TODO review naming between name and index
type Client struct {
	Id    UUID
	Index int
	conn  net.Conn
	raw   bool
}

func (c Client) handshake() error {
	c.conn.Write(EncodeString(gmInit, true))

	buf := make([]byte, 1024)
	if _, err := io.ReadAtLeast(c.conn, buf, 1); err != nil {
		return fmt.Errorf("socket read error %q", err.Error())
	}
	ackMagicNumber := uint32(binary.LittleEndian.Uint32(buf[0:4]))

	if ackMagicNumber != gmMagicNumber2 {
		return fmt.Errorf("incorrect handshake. Received: %d", ackMagicNumber)
	}
	log.Println("Finishing handshake")

	replyBuf := make([]byte, 12)
	binary.LittleEndian.PutUint32(replyBuf[0:4], gmMagicNumber3)
	binary.LittleEndian.PutUint32(replyBuf[4:8], gmMagicNumber4)
	binary.LittleEndian.PutUint32(replyBuf[8:12], gmMagicNumber5)
	if _, err := c.conn.Write(replyBuf); err != nil {
		return fmt.Errorf("error sending packet %s", err)
	}
	return nil
}

func (c Client) SendPacket(data []byte) {
	header := gmHeader(len(data))
	if !c.raw {
		if _, err := c.conn.Write(header); err != nil {
			log.Println("error sending packet:", err)
		}
	}
	if _, err := c.conn.Write(data); err != nil {
		log.Println("error sending packet:", err)
	}
}

// Alternative to SendPacket. Ideally we should probably manipulate
// buffer to make it easier to handle from client API.
func (c Client) SendBuffer(buffer *bytes.Buffer) {
	header := gmHeader(buffer.Len())
	if !c.raw {
		if _, err := c.conn.Write(header); err != nil {
			log.Println("error sending packet:", err)
		}
	}

	if _, err := buffer.WriteTo(c.conn); err != nil {
		log.Println("error writing to buffer")
	}
}

func gmHeader(dataLen int) []byte {
	buf := make([]byte, 12)
	binary.LittleEndian.PutUint32(buf[0:4], gmMagicNumber1)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(gmHeaderLen))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(dataLen))
	return buf
}

func NewClientId() UUID {
	return UUID(uuid.NewV4())
}
