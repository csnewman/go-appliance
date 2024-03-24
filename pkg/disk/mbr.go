package disk

import (
	"encoding/binary"
	"fmt"
	"strconv"

	rand "math/rand/v2"
)

const (
	MBRSize          = 512
	MBRPartitionSize = 16
	MBRSignature     = 0xAA55
)

type MBR struct {
	Bootstrap [440]byte
	DiskID    uint32
	Reserved  uint16
	Part1     MBRPartition
	Part2     MBRPartition
	Part3     MBRPartition
	Part4     MBRPartition
	Signature uint16
}

var CHSInvalid = CHSAddr{
	Head:     0xFF,
	Sector:   0xFF,
	Cylinder: 0xFF,
}

type CHSAddr struct {
	Head     byte
	Sector   byte
	Cylinder byte
}

func (a CHSAddr) String() string {
	return fmt.Sprintf("%v:%v:%v", a.Head, a.Sector, a.Cylinder)
}

type MBRPartType byte

const (
	MBRPartTypeLinux         = 0x83
	MBRPartTypeFAT32LBA      = 0x0C
	MBRPartTypeGPTProtective = 0xEE
)

func (t MBRPartType) String() string {
	switch t {
	case MBRPartTypeFAT32LBA:
		return "fat32lba"
	case MBRPartTypeLinux:
		return "linux"
	case MBRPartTypeGPTProtective:
		return "gpt-protective"
	default:
		return strconv.FormatUint(uint64(t), 16)
	}
}

type MBRPartition struct {
	Attrs    byte
	CHSStart CHSAddr
	Type     MBRPartType
	CHSLast  CHSAddr
	LBAStart uint32
	LBASize  uint32
}

func NewMBR() *MBR {
	return &MBR{
		Bootstrap: [440]byte{},
		DiskID:    rand.Uint32(),
		Reserved:  0,
		Part1:     MBRPartition{},
		Part2:     MBRPartition{},
		Part3:     MBRPartition{},
		Part4:     MBRPartition{},
		Signature: MBRSignature,
	}
}

func ParseMBR(data []byte) *MBR {
	if len(data) < MBRSize {
		panic("data < MBRSize")
	}

	return &MBR{
		Bootstrap: [440]byte(data[0:440]),
		DiskID:    binary.LittleEndian.Uint32(data[440:444]),
		Reserved:  binary.LittleEndian.Uint16(data[444:446]),
		Part1:     ParseMBRPartition(data[446:462]),
		Part2:     ParseMBRPartition(data[462:478]),
		Part3:     ParseMBRPartition(data[478:494]),
		Part4:     ParseMBRPartition(data[494:510]),
		Signature: binary.LittleEndian.Uint16(data[510:512]),
	}
}

func NewMBRPartition(ty MBRPartType, start uint32, size uint32) MBRPartition {
	return MBRPartition{
		Attrs:    0,
		CHSStart: CHSInvalid,
		Type:     ty,
		CHSLast:  CHSInvalid,
		LBAStart: start,
		LBASize:  size,
	}
}

func ParseMBRPartition(data []byte) MBRPartition {
	return MBRPartition{
		Attrs: data[0],
		CHSStart: CHSAddr{
			Head:     data[1],
			Sector:   data[2],
			Cylinder: data[3],
		},
		Type: MBRPartType(data[4]),
		CHSLast: CHSAddr{
			Head:     data[5],
			Sector:   data[6],
			Cylinder: data[7],
		},
		LBAStart: binary.LittleEndian.Uint32(data[8:12]),
		LBASize:  binary.LittleEndian.Uint32(data[12:16]),
	}
}

func (m *MBR) String() string {
	return fmt.Sprintf(
		"DiskID=%v Reserved=%v Part1={%v} Part2={%v} Part3={%v} Part4={%v} Signature=%v",
		m.DiskID,
		m.Reserved,
		m.Part1,
		m.Part2,
		m.Part3,
		m.Part4,
		m.Signature,
	)
}

func (p MBRPartition) String() string {
	if p == (MBRPartition{}) {
		return "empty"
	}

	return fmt.Sprintf(
		"Attrs=%v Type=%v CHSStart=%v CHSLast=%v LBAStart=%v LBASize=%v",
		p.Attrs,
		p.Type,
		p.CHSStart,
		p.CHSLast,
		p.LBAStart,
		p.LBASize,
	)
}

func (m *MBR) FillBytes(data []byte) {
	if len(data) < MBRSize {
		panic("data < MBRSize")
	}

	copy(data[0:440], m.Bootstrap[:])
	binary.LittleEndian.PutUint32(data[440:444], m.DiskID)
	binary.LittleEndian.PutUint16(data[444:447], m.Reserved)
	m.Part1.FillBytes(data[446:462])
	m.Part2.FillBytes(data[462:478])
	m.Part3.FillBytes(data[478:494])
	m.Part4.FillBytes(data[494:510])
	binary.LittleEndian.PutUint16(data[510:512], m.Signature)
}

func (p MBRPartition) FillBytes(data []byte) {
	if len(data) < MBRPartitionSize {
		panic("data < MBRPartitionSize")
	}

	data[0] = p.Attrs
	data[1] = p.CHSStart.Head
	data[2] = p.CHSStart.Sector
	data[3] = p.CHSStart.Cylinder
	data[4] = byte(p.Type)
	data[5] = p.CHSLast.Head
	data[6] = p.CHSLast.Sector
	data[7] = p.CHSLast.Cylinder
	binary.LittleEndian.PutUint32(data[8:12], p.LBAStart)
	binary.LittleEndian.PutUint32(data[12:16], p.LBASize)
}
