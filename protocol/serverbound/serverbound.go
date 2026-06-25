package serverbound

import "vexmc/protocol"

// handshake / toServer / types
type PacketHandshake struct {
	ProtocolVersion int32  `packet:"varint"`
	ServerAddress   string `packet:"string"`
	ServerPort      uint16 `packet:"ushort"`
	NextState       int32  `packet:"varint"`
}

type PacketStatusRequest struct{}
type PacketStatusPing struct {
	Time int64 `packet:"long"`
}

// login / toServer / types
type PacketLoginStart struct {
	Username string `packet:"string"`
}

// play / toServer / types
type PacketKeepAlive struct {
	KeepAliveID int32 `packet:"varint"`
}

type PacketChat struct {
	Message string `packet:"string"`
}

type PacketUseEntity struct {
	Target int32 `packet:"varint"`
	Mouse  int32 `packet:"varint"`
}

type PacketFlying struct {
	OnGround bool `packet:"boolean"`
}

type PacketPosition struct {
	X        float64 `packet:"double"`
	Y        float64 `packet:"double"`
	Z        float64 `packet:"double"`
	OnGround bool    `packet:"boolean"`
}

type PacketLook struct {
	Yaw      float32 `packet:"float"`
	Pitch    float32 `packet:"float"`
	OnGround bool    `packet:"boolean"`
}

type PacketPositionAndLook struct {
	X        float64 `packet:"double"`
	Y        float64 `packet:"double"`
	Z        float64 `packet:"double"`
	Yaw      float32 `packet:"float"`
	Pitch    float32 `packet:"float"`
	OnGround bool    `packet:"boolean"`
}

type PacketBlockDig struct {
	Status   int32             `packet:"varint"`
	Location protocol.Position `packet:"position"`
	Face     int8              `packet:"byte"`
}

type PacketBlockPlace struct {
	Location  protocol.Position `packet:"position"`
	Direction int8              `packet:"byte"`
	HeldItem  protocol.Slot     `packet:"slot"`
	CursorX   int8              `packet:"byte"`
	CursorY   int8              `packet:"byte"`
	CursorZ   int8              `packet:"byte"`
}

type PacketHeldItemSlot struct {
	SlotID int16 `packet:"short"`
}

type PacketEntityAction struct {
	EntityID  int32 `packet:"varint"`
	ActionID  int32 `packet:"varint"`
	JumpBoost int32 `packet:"varint"`
}

type PacketCloseWindow struct {
	WindowID uint8 `packet:"byte"`
}

type PacketWindowClick struct {
	WindowID    uint8         `packet:"byte"`
	Slot        int16         `packet:"short"`
	MouseButton int8          `packet:"byte"`
	Action      int16         `packet:"short"`
	Mode        int8          `packet:"byte"`
	Item        protocol.Slot `packet:"slot"`
}

type PacketTransaction struct {
	WindowID int8  `packet:"byte"`
	Action   int16 `packet:"short"`
	Accepted bool  `packet:"boolean"`
}

type PacketSetCreativeSlot struct {
	Slot int16         `packet:"short"`
	Item protocol.Slot `packet:"slot"`
}

type PacketEnchantItem struct {
	WindowID    int8 `packet:"byte"`
	Enchantment int8 `packet:"byte"`
}

type PacketUpdateSign struct {
	Location protocol.Position `packet:"position"`
	Text1    string            `packet:"string"`
	Text2    string            `packet:"string"`
	Text3    string            `packet:"string"`
	Text4    string            `packet:"string"`
}

type PacketAbilities struct {
	Flags        int8    `packet:"byte"`
	FlyingSpeed  float32 `packet:"float"`
	WalkingSpeed float32 `packet:"float"`
}
