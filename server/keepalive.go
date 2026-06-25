package server

import (
	"time"
	"vexmc/protocol"
	"vexmc/protocol/clientbound"
)

func (s *PlayerSession) KeepAlive() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		token := time.Now().UnixNano()
		_ = protocol.WritePacket(s.Conn, 0x00, clientbound.PacketKeepAlive{
			Time: int32(token),
		})
	}
}
