package protocol

import (
	"bytes"
	"errors"
	"io"
	"net"
)

func WriteVarInt(w io.Writer, val int32) error {
	uv := uint32(val)
	for {
		b := byte(uv & 0x7F)
		uv >>= 7
		if uv != 0 {
			b |= 0x80
		}
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
		if uv == 0 {
			break
		}
	}

	return nil
}

func ReadVarInt(r io.Reader) (int32, error) {
	var val int32
	var position int
	for {
		b := make([]byte, 1)
		if _, err := io.ReadFull(r, b); err != nil {
			return 0, err
		}
		val |= int32(b[0]&0x7F) << int32(position)
		if (b[0] & 0x80) == 0 {
			break
		}
		position += 7
		if position >= 32 {
			return 0, errors.New("VarInt too big")
		}
	}
	return val, nil
}

func ReadPacket(conn net.Conn) (int32, *bytes.Reader, error) {
	length, err := ReadVarInt(conn)
	if err != nil {
		return 0, nil, err
	}

	packetBuf := make([]byte, length)
	_, err = io.ReadFull(conn, packetBuf)
	if err != nil {
		return 0, nil, err
	}

	payloadReader := bytes.NewReader(packetBuf)
	packetID, err := ReadVarInt(payloadReader)

	if err != nil {
		return 0, nil, err
	}

	return packetID, payloadReader, nil
}

func WritePacket(conn net.Conn, packetID int32, packetStruct any) error {
	payload, err := MarshalPacket(packetStruct)
	if err != nil {
		return err
	}

	idBuf := new(bytes.Buffer)
	WriteVarInt(idBuf, packetID)
	totalLength := int32(idBuf.Len() + len(payload))

	if err := WriteVarInt(conn, totalLength); err != nil {
		return err
	}

	if _, err := conn.Write(idBuf.Bytes()); err != nil {
		return err
	}

	if _, err := conn.Write(payload); err != nil {
		return err
	}

	return nil
}
