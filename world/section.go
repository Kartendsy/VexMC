package world

const (
	SectionSize   = 16
	BlockDataSize = 8192
)

type Section struct {
	BlockData []byte
}

func NewSection() *Section {
	return &Section{
		BlockData: make([]byte, BlockDataSize),
	}
}

func (s *Section) SetBlock(x, y, z int, blockID uint16, meta byte) {
	idx := (y << 8) | (z << 4) | x
	val := (blockID << 4) | uint16(meta&15)

	s.BlockData[idx*2] = byte(val & 0xFF)
	s.BlockData[idx*2+1] = byte(val >> 8)
}
