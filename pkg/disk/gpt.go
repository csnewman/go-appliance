package disk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"unicode/utf16"

	"github.com/google/uuid"
)

const (
	GPTSize          = 92
	GPTPartitionSize = 128
	GPTVersion210    = 0x00010000
	GPTSignature     = 0x5452415020494645
)

var (
	GPTTypeMicrosoftBasicData = uuid.MustParse("EBD0A0A2-B9E5-4433-87C0-68B6B72699C7")
	GPTTypeLinuxFileSystem    = uuid.MustParse("0FC63DAF-8483-4772-8E79-3D69D8477DE4")
)

type GPT struct {
	Signature      uint64
	Revision       uint32
	Size           uint32
	Checksum       uint32
	Reserved       uint32
	ThisLBA        uint64
	AlternativeLBA uint64
	DataFirst      uint64
	DataLast       uint64
	GUID           uuid.UUID
	PartitionsLBA  uint64
	PartitionCount uint32
	EntrySize      uint32
	PartitionsCRC  uint32
}

var (
	ErrGPTNotPresent  = errors.New("GPT signature not detected")
	ErrGPTUnsupported = errors.New("GPT version unsupported")
)

func NewGPT(diskBlocks uint64) (*GPT, *GPT, error) {
	partBlocks := 32
	partCount := (BlockSize / GPTPartitionSize) * partBlocks

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, nil, err
	}

	primary := &GPT{
		Signature:      GPTSignature,
		Revision:       GPTVersion210,
		Size:           GPTSize,
		Checksum:       0,
		Reserved:       0,
		ThisLBA:        1,
		AlternativeLBA: diskBlocks - 1,
		DataFirst:      uint64(partBlocks + 2),
		DataLast:       diskBlocks - 2 - uint64(partCount),
		GUID:           id,
		PartitionsLBA:  2,
		PartitionCount: uint32(partCount),
		EntrySize:      GPTPartitionSize,
		PartitionsCRC:  0,
	}

	secondary := &GPT{
		Signature:      primary.Signature,
		Revision:       primary.Revision,
		Size:           primary.Size,
		Checksum:       0,
		Reserved:       primary.Reserved,
		ThisLBA:        primary.AlternativeLBA,
		AlternativeLBA: primary.ThisLBA,
		DataFirst:      primary.DataFirst,
		DataLast:       primary.DataLast,
		GUID:           primary.GUID,
		PartitionsLBA:  diskBlocks - 1 - uint64(partBlocks),
		PartitionCount: primary.PartitionCount,
		EntrySize:      primary.EntrySize,
		PartitionsCRC:  0,
	}

	return primary, secondary, nil
}

func ParseGPT(data []byte) (*GPT, error) {
	if len(data) < 16 {
		return nil, io.ErrUnexpectedEOF
	}

	sig := binary.LittleEndian.Uint64(data[0:8])

	if sig != GPTSignature {
		return nil, ErrGPTNotPresent
	}

	rev := binary.LittleEndian.Uint32(data[8:12])

	if rev != GPTVersion210 {
		return nil, fmt.Errorf("%w: revision %v", ErrGPTUnsupported, rev)
	}

	size := binary.LittleEndian.Uint32(data[12:16])

	if size != GPTSize {
		return nil, fmt.Errorf("%w: size %v", ErrGPTUnsupported, size)
	}

	if uint32(len(data)) < size {
		return nil, io.ErrUnexpectedEOF
	}

	cksum := binary.LittleEndian.Uint32(data[16:20])

	return &GPT{
		Signature:      sig,
		Revision:       rev,
		Size:           size,
		Checksum:       cksum,
		Reserved:       binary.LittleEndian.Uint32(data[20:24]),
		ThisLBA:        binary.LittleEndian.Uint64(data[24:32]),
		AlternativeLBA: binary.LittleEndian.Uint64(data[32:40]),
		DataFirst:      binary.LittleEndian.Uint64(data[40:48]),
		DataLast:       binary.LittleEndian.Uint64(data[48:56]),
		GUID:           guidFromBytes(data[56:72]),
		PartitionsLBA:  binary.LittleEndian.Uint64(data[72:80]),
		PartitionCount: binary.LittleEndian.Uint32(data[80:84]),
		EntrySize:      binary.LittleEndian.Uint32(data[84:88]),
		PartitionsCRC:  binary.LittleEndian.Uint32(data[88:92]),
	}, nil
}

func guidFromBytes(data []byte) uuid.UUID {
	var guid uuid.UUID

	guid[0] = data[3]
	guid[1] = data[2]
	guid[2] = data[1]
	guid[3] = data[0]

	guid[4] = data[5]
	guid[5] = data[4]

	guid[6] = data[7]
	guid[7] = data[6]

	copy(guid[8:16], data[8:16])

	return guid
}

func guidToBytes(guid uuid.UUID) []byte {
	var data [16]byte

	data[0] = guid[3]
	data[1] = guid[2]
	data[2] = guid[1]
	data[3] = guid[0]

	data[4] = guid[5]
	data[5] = guid[4]

	data[6] = guid[7]
	data[7] = guid[6]

	copy(data[8:16], guid[8:16])

	return data[:]
}

func (t *GPT) String() string {
	return fmt.Sprintf(
		"Signature=%v Revision=%v Size=%v Checksum=%v Reserved=%v ThisLBA=%v AlternativeLBA=%v DataFirst=%v DataLast=%v GUID=%v PartitionsLBA=%v PartitionCount=%v EntrySize=%v PartitionsCRC=%v", //nolint:lll
		t.Signature,
		t.Revision,
		t.Size,
		t.Checksum,
		t.Reserved,
		t.ThisLBA,
		t.AlternativeLBA,
		t.DataFirst,
		t.DataLast,
		t.GUID,
		t.PartitionsLBA,
		t.PartitionCount,
		t.EntrySize,
		t.PartitionsCRC,
	)
}

func (t *GPT) FillBytes(data []byte) {
	if uint32(len(data)) < t.Size {
		panic("data < gpt header size")
	}

	if uint32(len(data)) < GPTSize {
		panic("data < GPTSize")
	}

	binary.LittleEndian.PutUint64(data[0:8], t.Signature)
	binary.LittleEndian.PutUint32(data[8:12], t.Revision)
	binary.LittleEndian.PutUint32(data[12:16], t.Size)
	binary.LittleEndian.PutUint32(data[16:20], t.Checksum)
	binary.LittleEndian.PutUint32(data[20:24], t.Reserved)
	binary.LittleEndian.PutUint64(data[24:32], t.ThisLBA)
	binary.LittleEndian.PutUint64(data[32:40], t.AlternativeLBA)
	binary.LittleEndian.PutUint64(data[40:48], t.DataFirst)
	binary.LittleEndian.PutUint64(data[48:56], t.DataLast)
	copy(data[56:72], guidToBytes(t.GUID))
	binary.LittleEndian.PutUint64(data[72:80], t.PartitionsLBA)
	binary.LittleEndian.PutUint32(data[80:84], t.PartitionCount)
	binary.LittleEndian.PutUint32(data[84:88], t.EntrySize)
	binary.LittleEndian.PutUint32(data[88:92], t.PartitionsCRC)
}

func (t *GPT) CalculateChecksum() uint32 {
	var data [GPTSize]byte

	t.FillBytes(data[:])

	// Clear checksum
	data[16] = 0
	data[17] = 0
	data[18] = 0
	data[19] = 0

	return crc32.ChecksumIEEE(data[:])
}

type GPTPartition struct {
	Type       uuid.UUID
	ID         uuid.UUID
	StartLBA   uint64
	EndLBA     uint64
	Attributes uint64
	Name       string
}

func NewGPTPartition(ty uuid.UUID, start uint64, end uint64, name string) (GPTPartition, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return GPTPartition{}, err
	}

	return GPTPartition{
		Type:       ty,
		ID:         id,
		StartLBA:   start,
		EndLBA:     end,
		Attributes: 0,
		Name:       name,
	}, nil
}

func ParseGPTPartitions(reader io.ReaderAt, start uint64, size uint32, count uint32) ([]GPTPartition, uint32, error) {
	if size < GPTPartitionSize {
		panic("size < GPTPartitionSize")
	}

	parts := make([]GPTPartition, count)
	data := make([]byte, size)
	chars := make([]uint16, (size-56)/2)

	hasher := crc32.NewIEEE()

	for i := uint32(0); i < count; i++ {
		_, err := reader.ReadAt(data, int64(start))
		if err != nil {
			return nil, 0, err
		}

		_, err = hasher.Write(data)
		if err != nil {
			return nil, 0, err
		}

		nameLen := 0

		for ; nameLen < len(chars); nameLen++ {
			pos := 56 + (nameLen * 2)

			chars[nameLen] = binary.LittleEndian.Uint16(data[pos : pos+2])

			if chars[nameLen] == 0 {
				break
			}
		}

		name := ""

		if nameLen > 0 {
			name = string(utf16.Decode(chars[:nameLen]))
		}

		parts[i] = GPTPartition{
			Type:       guidFromBytes(data[0:16]),
			ID:         guidFromBytes(data[16:32]),
			StartLBA:   binary.LittleEndian.Uint64(data[32:40]),
			EndLBA:     binary.LittleEndian.Uint64(data[40:48]),
			Attributes: binary.LittleEndian.Uint64(data[48:56]),
			Name:       name,
		}

		start += uint64(size)
	}

	return parts, hasher.Sum32(), nil
}

func WriteGPTPartitions(writer io.WriterAt, start uint64, size uint32, parts []GPTPartition) (uint32, error) {
	if size < GPTPartitionSize {
		panic("size < GPTPartitionSize")
	}

	hasher := crc32.NewIEEE()

	data := make([]byte, size)

	for _, part := range parts {
		clear(data)

		copy(data[0:16], guidToBytes(part.Type))
		copy(data[16:32], guidToBytes(part.ID))
		binary.LittleEndian.PutUint64(data[32:40], part.StartLBA)
		binary.LittleEndian.PutUint64(data[40:48], part.EndLBA)
		binary.LittleEndian.PutUint64(data[48:56], part.Attributes)

		if len(part.Name) > 0 {
			encoded := utf16.Encode([]rune(part.Name))

			for i, v := range encoded {
				i = i*2 + 56

				binary.LittleEndian.PutUint16(data[i:i+2], v)
			}
		}

		_, err := hasher.Write(data)
		if err != nil {
			return 0, err
		}

		_, err = writer.WriteAt(data, int64(start))
		if err != nil {
			return 0, err
		}

		start += uint64(size)
	}

	return hasher.Sum32(), nil
}
