package world

import (
	"fmt"
	"sync"
)

type WorldManager struct {
	chunks map[string]*Chunk
	sync.RWMutex
}

func NewWorldManager() *WorldManager {
	return &WorldManager{
		chunks: make(map[string]*Chunk),
	}
}

func (wm *WorldManager) getChunkKey(x, z int) string {
	return fmt.Sprintf("%d,%d", x, z)
}

func (wm *WorldManager) GetChunk(x, z int) *Chunk {
	wm.Lock()
	defer wm.Unlock()

	key := wm.getChunkKey(x, z)
	chunk, exists := wm.chunks[key]

	if !exists {
		chunk = wm.generateDefaultChunk(x, z)
		wm.chunks[key] = chunk
	}
	return chunk
}

func (wm *WorldManager) generateDefaultChunk(x, z int) *Chunk {
	chunk := NewChunk(x, z)

	for i := range chunk.Biomes {
		chunk.Biomes[i] = 1
	}

	for cx := range 16 {
		for cz := range 16 {
			chunk.SetBlock(cx, 0, cz, 7, 0)
			chunk.SetBlock(cx, 1, cz, 3, 0)
			chunk.SetBlock(cx, 2, cz, 3, 0)
			chunk.SetBlock(cx, 3, cz, 2, 0)
		}
	}

	return chunk
}

func (wm *WorldManager) UnloadChunk(x, z int) {
	wm.Lock()
	defer wm.Unlock()

	key := wm.getChunkKey(x, z)
	delete(wm.chunks, key)
}
