package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var mbrRaspPiRaw = [MBRSize]byte{
	250, 184, 0, 16, 142, 208, 188, 0, 176, 184, 0, 0, 142, 216, 142, 192, 251, 190, 0, 124, 191, 0, 6, 185, 0, 2, 243,
	164, 234, 33, 6, 0, 0, 190, 190, 7, 56, 4, 117, 11, 131, 198, 16, 129, 254, 254, 7, 117, 243, 235, 22, 180, 2, 176,
	1, 187, 0, 124, 178, 128, 138, 116, 1, 139, 76, 2, 205, 19, 234, 0, 124, 0, 0, 235, 254, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 154, 72, 164, 26, 0, 0, 0, 0, 1, 64, 12, 3, 224, 255, 0, 32, 0, 0, 0, 0, 16, 0, 0, 3, 224, 255, 131,
	3, 224, 255, 0, 32, 16, 0, 0, 160, 145, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 85, 170,
}

var mbrRaspPiParsed = &MBR{
	Bootstrap: [440]byte{
		250, 184, 0, 16, 142, 208, 188, 0, 176, 184, 0, 0, 142, 216, 142, 192, 251, 190, 0, 124, 191, 0, 6, 185, 0, 2,
		243, 164, 234, 33, 6, 0, 0, 190, 190, 7, 56, 4, 117, 11, 131, 198, 16, 129, 254, 254, 7, 117, 243, 235, 22, 180,
		2, 176, 1, 187, 0, 124, 178, 128, 138, 116, 1, 139, 76, 2, 205, 19, 234, 0, 124, 0, 0, 235, 254,
	},
	DiskID:   446974106,
	Reserved: 0,
	Part1: MBRPartition{
		Attrs: 0,
		CHSStart: CHSAddr{
			Head:     0,
			Sector:   1,
			Cylinder: 64,
		},
		Type: MBRPartTypeFAT32LBA,
		CHSLast: CHSAddr{
			Head:     3,
			Sector:   224,
			Cylinder: 255,
		},
		LBAStart: 8192,
		LBASize:  1048576,
	},
	Part2: MBRPartition{
		Attrs: 0,
		CHSStart: CHSAddr{
			Head:     3,
			Sector:   224,
			Cylinder: 255,
		},
		Type: MBRPartTypeLinux,
		CHSLast: CHSAddr{
			Head:     3,
			Sector:   224,
			Cylinder: 255,
		},
		LBAStart: 1056768,
		LBASize:  9543680,
	},
	Part3:     MBRPartition{},
	Part4:     MBRPartition{},
	Signature: 43605,
}

func TestMBR(t *testing.T) {
	parsed := ParseMBR(mbrRaspPiRaw[:])
	assert.Equal(t, mbrRaspPiParsed, parsed, "parsed mbr should match")

	var encoded [MBRSize]byte

	parsed.FillBytes(encoded[:])

	assert.Equal(t, mbrRaspPiRaw, encoded, "re-encoded mbr should match")
}
