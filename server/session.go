package server

import (
	"net"
	"sync"
	"vexmc/protocol"
	"vexmc/protocol/clientbound"
	"vexmc/world"
)

type PlayerSession struct {
	Conn     net.Conn
	Username string
	UUID     string
	X, Y, Z  float64
	OnGround bool

	LastChunkX   int
	LastChunkZ   int
	ViewDistance int
}

func (p *PlayerSession) UpdatePosition(newX, newZ float64, worldMap *world.WorldManager) {
	currentChunkX := int(newX) >> 4
	currentChunkZ := int(newZ) >> 4

	if currentChunkX == p.LastChunkX && currentChunkZ == p.LastChunkZ {
		return
	}

	r := p.ViewDistance

	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			targetX := currentChunkX + x
			targetZ := currentChunkZ + z

			if !inRadius(targetX, targetZ, p.LastChunkX, p.LastChunkZ, r) {
				chunk := worldMap.GetChunk(targetX, targetZ)
				data, bitmask := chunk.Serialize()

				_ = protocol.WritePacket(p.Conn, 0x21, clientbound.PacketMapChunk{
					ChunkX:   int32(targetX),
					ChunkZ:   int32(targetZ),
					GroundUp: true,
					BitMap:   bitmask,
					Data:     data,
				})
			}
		}
	}

	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			oldX := p.LastChunkX + x
			oldZ := p.LastChunkZ + z

			if !inRadius(oldX, oldZ, currentChunkX, currentChunkZ, r) {
				_ = protocol.WritePacket(p.Conn, 0x21, clientbound.PacketMapChunk{
					ChunkX:   int32(oldX),
					ChunkZ:   int32(oldZ),
					GroundUp: true,
					BitMap:   0,
					Data:     []byte{},
				})
			}
		}
	}

	p.LastChunkX = currentChunkX
	p.LastChunkZ = currentChunkZ
}

func inRadius(targetX, targetZ, centerX, centerZ, radius int) bool {
	return targetX >= centerX-radius && targetX <= centerX+radius && targetZ >= centerZ-radius && targetZ <= centerZ+radius
}

type SessionManager struct {
	sync.RWMutex
	players map[string]*PlayerSession
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		players: make(map[string]*PlayerSession),
	}
}

func (sm *SessionManager) Add(uuid string, session *PlayerSession) {
	sm.Lock()
	defer sm.Unlock()
	sm.players[uuid] = session
}

func (sm *SessionManager) Remove(uuid string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.players, uuid)
}
