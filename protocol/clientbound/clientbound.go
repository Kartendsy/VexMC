package clientbound

type PacketStatusResponse struct {
	JSONResponse string `packet:"string"`
}

type PacketStatusPong struct {
	Time int64 `packet:"long"`
}

// login / toClient / types
type PacketLoginSuccess struct {
	UUID     string `packet:"string"`
	Username string `packet:"string"`
}

// play / toClient / types
type PacketKeepAlive struct {
	Time int32 `packet:"varint"`
}

type PacketJoinGame struct {
	EntityID         int32  `packet:"int"`
	GameMode         uint8  `packet:"byte"`
	Dimension        uint8  `packet:"byte"`
	Difficulty       uint8  `packet:"byte"`
	MaxPlayers       uint8  `packet:"byte"`
	LevelType        string `packet:"string"`
	ReducedDebugInfo bool   `packet:"boolean"`
}

type PacketChat struct {
	Message  string `packet:"string"`
	Position uint8  `packet:"byte"`
}

type PacketPosition struct {
	X     float64 `packet:"double"`
	Y     float64 `packet:"double"`
	Z     float64 `packet:"double"`
	Yaw   float32 `packet:"float"`
	Pitch float32 `packet:"float"`
	Flags uint8   `packet:"byte"`
}

type PacketMapChunk struct {
	ChunkX   int32  `packet:"int"`
	ChunkZ   int32  `packet:"int"`
	GroundUp bool   `packet:"boolean"`
	BitMap   uint16 `packet:"ushort"`
	Data     []byte `packet:"bytearray"`
}
