package shm

import (
	"encoding/binary"
)

type FrameMetadata struct {
	Timestamp  uint64
	Magic      uint32
	FrameSize  uint32
	CodecID    uint32
	IsKeyFrame uint8
	MediaType  uint8
}

func (fm *FrameMetadata) GetMagic(metaBytes []byte) (uint32) {
	return binary.LittleEndian.Uint32(metaBytes[8:12])
}

func (fm *FrameMetadata) LoadFrom(metaBytes []byte) {

	fm.Timestamp =  binary.LittleEndian.Uint64(metaBytes[0:8])
	fm.Magic      = binary.LittleEndian.Uint32(metaBytes[8:12])
	fm.FrameSize =  binary.LittleEndian.Uint32(metaBytes[12:16])
	fm.CodecID =    binary.LittleEndian.Uint32(metaBytes[16:20])
	fm.IsKeyFrame = metaBytes[20]
	fm.MediaType =  metaBytes[21]

}
