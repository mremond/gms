package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/mremond/gamemaker-server"
)

const (
	gmProtocolRaw = true
	gmPort        = 9000
)

// Message implemented in the server:
const (
	msgMove = 1 + iota
)

// TODO Maybe move this map in server and grant client access to this map
var clients map[gms.UUID]gms.Client

func init() {
	clients = make(map[gms.UUID]gms.Client)
}

func main() {
	server := gms.Server{Raw: gmProtocolRaw}
	server.Start(gmPort, HandleEvent)
}

// GM Event type dispatcher
// if HandleEvent return an error, the server will as a result disconnect that client.
func HandleEvent(msg gms.Message) error {
	switch msg.EventType {
	case gms.ClientConnect:
		return handleConnect(msg)
	case gms.ClientData:
		return HandleData(msg)
	case gms.ClientDisconnect:
		handleDisconnect(msg)
	}
	return nil
}

func handleConnect(msg gms.Message) error {
	fmt.Println("Client connected:", msg.Client.Id)

	clients[msg.Client.Id] = msg.Client

	return nil
}

func handleDisconnect(msg gms.Message) error {
	fmt.Println("Client disconnected:", msg.Client.Id)

	delete(clients, msg.Client.Id)

	return nil
}

func HandleData(msg gms.Message) error {
	switch msg.DataType {
	case msgMove:
		// fmt.Println("Received move event")
		x, y, _ := readMove(msg.Buffer)
		//fmt.Println("move to:", x, y)
		for _, client := range clients {
			if client.Id != msg.Client.Id {
				sendMove(client, x, y)
			}
		}
	}
	return nil
}

func readMove(buffer gms.Reader) (uint16, uint16, error) {
	x, err := gms.ReadUint16(buffer)
	if err != nil {
		return x, 0, err
	}
	y, err := gms.ReadUint16(buffer)
	return x, y, err
}

func sendMove(c gms.Client, x, y uint16) error {
	fmt.Println("Send move for id:", uint16(c.Index))
	buf := make([]byte, 7)
	buf[0] = msgMove
	binary.LittleEndian.PutUint16(buf[1:3], uint16(c.Index))
	binary.LittleEndian.PutUint16(buf[3:5], x)
	binary.LittleEndian.PutUint16(buf[5:7], y)
	c.SendPacket(buf)
	return nil
}

// Does not work.
func sendMove2(c gms.Client, x, y uint16) error {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, uint8(msgMove))
	binary.Write(b, binary.LittleEndian, uint16(c.Index))
	binary.Write(b, binary.LittleEndian, x)
	binary.Write(b, binary.LittleEndian, y)
	c.SendBuffer(b)
	return nil
}
