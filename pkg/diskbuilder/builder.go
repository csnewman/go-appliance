package diskbuilder

import (
	"errors"
	"fmt"

	"github.com/csnewman/go-appliance/pkg/disk"
)

var ErrInvalidSize = errors.New("invalid disk size")

type Builder struct {
	Disk        *disk.Disk
	MBR         *disk.MBR
	Primary     *disk.GPT
	Secondary   *disk.GPT
	Parts       []disk.GPTPartition
	LastPart    int
	LastMBRPart int
}

func New(path string, size int64) (*Builder, error) {
	if size%disk.BlockSize != 0 {
		return nil, ErrInvalidSize
	}

	blocks := size / disk.BlockSize

	d, err := disk.Create(path, size)
	if err != nil {
		return nil, err
	}

	mbr := disk.NewMBR()

	primary, secondary, err := disk.NewGPT(uint64(blocks))
	if err != nil {
		return nil, fmt.Errorf("failed to create gpt table: %w", err)
	}

	parts := make([]disk.GPTPartition, primary.PartitionCount)

	return &Builder{
		Disk:      d,
		MBR:       mbr,
		Parts:     parts,
		Primary:   primary,
		Secondary: secondary,
	}, nil
}

func (b *Builder) Add(gpt disk.GPTPartition) {
	b.Parts[b.LastPart] = gpt
	b.LastPart++
}

func (b *Builder) AddMBR(mbr disk.MBRPartition) {
	switch b.LastMBRPart {
	case 0:
		b.MBR.Part1 = mbr
	case 1:
		b.MBR.Part2 = mbr
	case 2:
		b.MBR.Part3 = mbr
	case 3:
		b.MBR.Part4 = mbr
	default:
		panic("only 4 MBR partitions allowed")
	}

	b.LastMBRPart++
}

func (b *Builder) Close() error {
	if err := b.Disk.WriteMBR(b.MBR); err != nil {
		return fmt.Errorf("faile to write MBR: %w", err)
	}

	partCrc, err := b.Disk.WriteGPTPartitions(
		b.Primary.PartitionsLBA*disk.BlockSize,
		b.Primary.EntrySize,
		b.Parts,
	)
	if err != nil {
		return fmt.Errorf("failed to write primary parts: %w", err)
	}

	b.Primary.PartitionsCRC = partCrc
	b.Primary.Checksum = b.Primary.CalculateChecksum()

	if err := b.Disk.WriteGPT(b.Primary.ThisLBA, b.Primary); err != nil {
		return fmt.Errorf("failed to write primary gpt: %w", err)
	}

	partCrc, err = b.Disk.WriteGPTPartitions(
		b.Secondary.PartitionsLBA*disk.BlockSize,
		b.Secondary.EntrySize,
		b.Parts,
	)
	if err != nil {
		return fmt.Errorf("failed to write secondary parts: %w", err)
	}

	b.Secondary.PartitionsCRC = partCrc
	b.Secondary.Checksum = b.Secondary.CalculateChecksum()

	if err := b.Disk.WriteGPT(b.Secondary.ThisLBA, b.Secondary); err != nil {
		return fmt.Errorf("failed to write secondary gpt: %w", err)
	}

	return nil
}
