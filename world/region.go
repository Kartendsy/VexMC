package world

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"os"
	"time"
)

const SectorSize = 4096

type Region struct {
	file       *os.File
	offsets    [1024]uint32
	timestamps [1024]uint32
}

func OpenRegion(path string) (*Region, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	r := &Region{file: file}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fi.Size() < SectorSize*2 {
		emptyHeader := make([]byte, SectorSize*2)
		if _, err := file.Write(emptyHeader); err != nil {
			return nil, err
		}
	} else {
		header := make([]byte, SectorSize*2)
		if _, err := file.ReadAt(header, 0); err != nil {
			return nil, err
		}
		for i := range 1024 {
			r.offsets[i] = binary.BigEndian.Uint32(header[i*4 : i*4+4])
			r.timestamps[i] = binary.BigEndian.Uint32(header[SectorSize+i*4 : SectorSize+i*4+4])
		}
	}

	return r, nil
}

func (r *Region) Close() error {
	return r.file.Close()
}

func (r *Region) SaveChunk(c *Chunk) error {
	localX := c.X % 32
	if localX < 0 {
		localX += 32
	}
	localZ := c.Z % 32
	if localZ < 0 {
		localZ += 32
	}

	tableIdx := localX + localZ*32

	rawChunkData, bitmask := c.Serialize()

	payloadBuf := new(bytes.Buffer)
	binary.Write(payloadBuf, binary.BigEndian, bitmask)
	payloadBuf.Write(rawChunkData)

	var compressedBuf bytes.Buffer
	zlibw := zlib.NewWriter(&compressedBuf)
	if _, err := zlibw.Write(payloadBuf.Bytes()); err != nil {
		return err
	}
	zlibw.Close()

	chunkPayloadLength := uint32(compressedBuf.Len() + 1)

	finalChunkBytes := new(bytes.Buffer)
	binary.Write(finalChunkBytes, binary.BigEndian, chunkPayloadLength)
	finalChunkBytes.WriteByte(2)
	finalChunkBytes.Write(compressedBuf.Bytes())

	totalBytes := finalChunkBytes.Len()
	sectorsNeeded := (totalBytes + SectorSize - 1) / SectorSize

	padding := (sectorsNeeded * SectorSize) - totalBytes
	finalChunkBytes.Write(make([]byte, padding))

	fi, err := r.file.Stat()
	if err != nil {
		return err
	}

	offsetSector := uint32(fi.Size() / SectorSize)

	if _, err := r.file.WriteAt(finalChunkBytes.Bytes(), int64(offsetSector*SectorSize)); err != nil {
		return err
	}

	r.offsets[tableIdx] = (offsetSector << 8) | uint32(sectorsNeeded&0xFF)
	r.timestamps[tableIdx] = uint32(time.Now().Unix())

	headerBuf := make([]byte, SectorSize*2)
	for i := range 1024 {
		binary.BigEndian.PutUint32(headerBuf[i*4:], r.offsets[i])
		binary.BigEndian.PutUint32(headerBuf[SectorSize+i*4:], r.timestamps[i])
	}

	_, err = r.file.WriteAt(headerBuf, 0)
	return err

}
