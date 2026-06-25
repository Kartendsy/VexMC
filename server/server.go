package server

import (
	"fmt"
	"net"
	"reflect"
	"vexmc/logger"
	"vexmc/protocol"
	"vexmc/protocol/clientbound"
	"vexmc/protocol/serverbound"
	"vexmc/world"

	"github.com/google/uuid"
)

type Server struct {
	SessionManager *SessionManager
	WorldManager   *world.WorldManager
}

func NewServer() *Server {
	s := &Server{
		SessionManager: NewSessionManager(),
		WorldManager:   world.NewWorldManager(),
	}

	return s
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", ":25565")
	if err != nil {
		return
	}

	defer ln.Close()

	logger.Info("Server is running on 0.0.0.0:25565")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go s.handshake(conn)
	}
}

func (s *Server) handshake(conn net.Conn) {
	packetID, payload, err := protocol.ReadPacket(conn)
	if err != nil {
		return
	}

	if packetID != 0x00 {
		return
	}

	var handshake serverbound.PacketHandshake
	protocol.UnmarshalPacket(payload, &handshake)

	switch handshake.NextState {
	case 1:
		s.status(conn)
	case 2:
		s.login(conn)
	}

}

func (s *Server) status(conn net.Conn) {
	for {
		packetID, payload, err := protocol.ReadPacket(conn)
		if err != nil {
			return
		}

		switch packetID {
		case 0x00:
			motdJSON := `{
								"version": { "name": "Go-MC 1.8.8", "protocol": 47 },
								"players": { "max": 2026, "online": 0},
								"description": { "text": "§a§lKoneksi TCP Sukses! §r\n§eDibuat menggunakan Golang Struct" }
							}`
			response := clientbound.PacketStatusResponse{JSONResponse: motdJSON}
			protocol.WritePacket(conn, 0x00, response)
		case 0x01:
			var ping serverbound.PacketStatusPing
			protocol.UnmarshalPacket(payload, &ping)

			pong := clientbound.PacketStatusPong{Time: ping.Time}
			protocol.WritePacket(conn, 0x01, pong)
			return
		}
	}

}

func (s *Server) login(conn net.Conn) {
	packetID, payload, err := protocol.ReadPacket(conn)
	if err != nil || packetID != 0x00 {
		return
	}

	var loginStart serverbound.PacketLoginStart
	protocol.UnmarshalPacket(payload, &loginStart)

	sessionUUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte(loginStart.Username))

	loginSuccess := clientbound.PacketLoginSuccess{
		UUID:     sessionUUID.String(),
		Username: loginStart.Username,
	}

	_ = protocol.WritePacket(conn, 0x02, loginSuccess)

	session := &PlayerSession{
		Conn:         conn,
		Username:     loginStart.Username,
		UUID:         sessionUUID.String(),
		X:            0,
		Y:            64,
		Z:            0,
		LastChunkX:   0,
		LastChunkZ:   0,
		ViewDistance: 4,
	}

	_ = protocol.WritePacket(conn, 0x01, clientbound.PacketJoinGame{
		EntityID:         12345,
		GameMode:         1,
		Dimension:        0,
		Difficulty:       1,
		MaxPlayers:       10,
		LevelType:        "flat",
		ReducedDebugInfo: false,
	})

	StreamInitialChunks(session.Conn, 0, 0, session.ViewDistance, s.WorldManager)

	_ = protocol.WritePacket(conn, 0x08, clientbound.PacketPosition{
		X:     0,
		Y:     64,
		Z:     0,
		Yaw:   0,
		Pitch: 0,
		Flags: 1,
	})

	logger.Info("%s Joined server...", loginStart.Username)

	s.SessionManager.Add(loginSuccess.UUID, session)

	if method := reflect.ValueOf(session).MethodByName("KeepAlive"); method.IsValid() {
		go session.KeepAlive()
	}

	s.play(session)
}

func (s *Server) play(session *PlayerSession) {
	defer func() {
		session.Conn.Close()
		logger.Info("%s Leave server...", session.Username)
		s.SessionManager.Remove(session.UUID)
	}()

	for {
		packetID, payload, err := protocol.ReadPacket(session.Conn)
		if err != nil {
			return
		}

		switch packetID {
		case 0x01:
			var chat serverbound.PacketChat
			_ = protocol.UnmarshalPacket(payload, &chat)
			fmt.Printf("[%s]: %s", session.Username, chat.Message)

			jsonStr := fmt.Sprintf(`{"text":"<%s> %s"}`, session.Username, chat.Message)
			_ = protocol.WritePacket(session.Conn, 0x02, clientbound.PacketChat{
				Message:  jsonStr,
				Position: 0,
			})

		case 0x04:
			var pos serverbound.PacketPosition
			err := protocol.UnmarshalPacket(payload, &pos)
			if err != nil {
				fmt.Println("Gagal membaca packet posisi: ", err)
				continue
			}
			session.UpdatePosition(pos.X, pos.Z, s.WorldManager)

		case 0x12:
			var updateSign serverbound.PacketUpdateSign
			if err := protocol.UnmarshalPacket(payload, &updateSign); err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("\n Text1: %s \n Text2: %s \n Text3: %s \n Text4: %s", updateSign.Text1, updateSign.Text2, updateSign.Text3, updateSign.Text4)
		}
	}
}

func StreamInitialChunks(conn net.Conn, pX, pZ float64, viewDistance int, worldMap *world.WorldManager) {
	centerX := int(pX) >> 4
	centerZ := int(pZ) >> 4

	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			targetX := centerX + x
			targetZ := centerZ + z

			chunk := worldMap.GetChunk(targetX, targetZ)

			data, bitmask := chunk.Serialize()

			_ = protocol.WritePacket(conn, 0x21, clientbound.PacketMapChunk{
				ChunkX:   int32(targetX),
				ChunkZ:   int32(targetZ),
				GroundUp: true,
				BitMap:   bitmask,
				Data:     data,
			})
		}
	}
}
