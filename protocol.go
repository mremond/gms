package gms

import (
	"bytes"
	"encoding/binary"
)

// EncodeString encodes a null terminated string.
func EncodeString(str string, nullTerminated bool) []byte {
	if !nullTerminated {
		length := make([]byte, 2)
		binary.LittleEndian.PutUint16(length, uint16(len(str)))
		return append(length, []byte(str)...)
	}
	return append([]byte(str), []byte{0}...)
}

// ReadString returns the first null terminated string from byte slice.
func ReadString(buffer Reader) (string, error) {
	var result bytes.Buffer
	for {
		data, err := buffer.Read(1)
		if err != nil {
			return "", err
		}
		if data[0] == '\x00' {
			break
		}
		result.WriteByte(data[0])
	}
	return result.String(), nil
}

// ReadUint8 reads uint8 from buffer.
func ReadUint8(buffer Reader) (uint8, error) {
	data, err := buffer.Read(1)
	if err != nil {
		return 0, err
	}
	return uint8(data[0]), err
}

// ReadUint16 reads uint16 from buffer.
func ReadUint16(buffer Reader) (uint16, error) {
	data, err := buffer.Read(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), err
}

// ReadUint32 reads uint32 from buffer.
func ReadUint32(buffer Reader) (uint32, error) {
	data, err := buffer.Read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), err
}
