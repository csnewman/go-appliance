package disk

import (
	"crypto/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPTGuid(t *testing.T) {
	bytes := []byte{0x28, 0x73, 0x2a, 0xc1, 0x1f, 0xf8, 0xd2, 0x11, 0xba, 0x4b, 0x00, 0xa0, 0xc9, 0x3e, 0xc9, 0x3b}
	expected := uuid.MustParse("C12A7328-F81F-11D2-BA4B-00A0C93EC93B")

	fromBytes := guidFromBytes(bytes)

	assert.Equal(t, expected, fromBytes, "parsed guid should match")

	toBytes := guidToBytes(expected)

	assert.Equal(t, bytes, toBytes, "encoded bytes should match")
}

type ByteBuf struct {
	buf []byte
}

func (b *ByteBuf) ReadAt(data []byte, off int64) (int, error) {
	copy(data, b.buf[off:])

	return len(data), nil
}

func (b *ByteBuf) WriteAt(data []byte, off int64) (int, error) {
	copy(b.buf[off:], data)

	return len(data), nil
}

func TestGPTPartitions(t *testing.T) {
	data := make([]byte, 512*1024)
	buf := &ByteBuf{buf: data}

	_, err := rand.Read(data)

	require.NoError(t, err, "rand data")

	parts := []GPTPartition{
		{
			Type:       uuid.MustParse("389b2069-8a04-4768-b803-5bfb85a79054"),
			ID:         uuid.MustParse("fadad17f-f91d-45ef-a67b-df68943a43fb"),
			StartLBA:   1234,
			EndLBA:     4567,
			Attributes: 6543,
			Name:       "Part 1",
		},
		{
			Type:       uuid.MustParse("e71f10ed-2997-4852-92a7-a9ac222fddc8"),
			ID:         uuid.MustParse("25cc653a-5090-48f5-956f-faf2867991c4"),
			StartLBA:   1111,
			EndLBA:     2222,
			Attributes: 3333,
			Name:       "Another",
		},
		{},
		{
			Type:       uuid.MustParse("525fbf67-cf11-4a1d-87aa-a6616d6a5b70"),
			ID:         uuid.MustParse("9d3b824b-792b-4b09-8c05-77a8c76dbe38"),
			StartLBA:   1234567890,
			EndLBA:     1234567890,
			Attributes: 1234567890,
			Name:       "Final One",
		},
	}

	crc, err := WriteGPTPartitions(buf, 1234, 128, parts)

	require.NoError(t, err, "parts should write")
	assert.Equal(t, uint32(0xf9937558), crc, "crc should match")

	readParts, crc, err := ParseGPTPartitions(buf, 1234, 128, uint32(len(parts)))

	require.NoError(t, err, "parts should read")
	assert.Equal(t, uint32(0xf9937558), crc, "crc should match")
	assert.Equal(t, parts, readParts, "reread parts should match")
}
