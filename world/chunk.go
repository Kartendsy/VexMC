package world

import "bytes"

type Chunk struct {
	X, Z     int
	Sections [16]*Section
	Biomes   []byte
}

func NewChunk(x, z int) *Chunk {
	return &Chunk{
		X:      x,
		Z:      z,
		Biomes: make([]byte, 256),
	}
}

func (c *Chunk) SetBlock(x, y, z int, blockID uint16, meta byte) {
	if y < 0 || y > 255 || x < 0 || x > 15 || z < 0 || z > 15 {
		return
	}

	secIdx := y / 16
	localY := y % 16

	if c.Sections[secIdx] == nil {
		c.Sections[secIdx] = NewSection()
	}

	c.Sections[secIdx].SetBlock(x, localY, z, blockID, meta)
}

func (c *Chunk) Serialize() ([]byte, uint16) {
	var bitmask uint16 = 0
	activeSectionsCount := 0

	for i := range 16 {
		if c.Sections[i] != nil {
			bitmask |= (1 << i)
			activeSectionsCount++
		}
	}

	totalSize := (BlockDataSize + 2048 + 2048) * activeSectionsCount
	finalData := make([]byte, 0, totalSize)

	for i := range 16 {
		if c.Sections[i] != nil {
			finalData = append(finalData, c.Sections[i].BlockData...)
		}
	}

	if activeSectionsCount > 0 {
		lightSize := 2048 * activeSectionsCount
		finalData = append(finalData, bytes.Repeat([]byte{0xFF}, lightSize)...)
		finalData = append(finalData, bytes.Repeat([]byte{0xFF}, lightSize)...)
	}

	finalData = append(finalData, c.Biomes...)

	return finalData, bitmask
}
