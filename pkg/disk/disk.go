package disk

import (
	"fmt"
	"io"
	"os"
)

const BlockSize = 512

type Disk struct {
	file *os.File
}

func Open(dev string) (*Disk, error) {
	f, err := os.OpenFile(dev, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &Disk{
		file: f,
	}, nil
}

func Create(dev string, size int64) (*Disk, error) {
	f, err := os.Create(dev)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	if err := f.Truncate(size); err != nil {
		return nil, fmt.Errorf("failed to resize file: %w", err)
	}

	return &Disk{
		file: f,
	}, nil
}

func (d *Disk) Close() error {
	return d.file.Close()
}

func (d *Disk) ReadMBR() (*MBR, error) {
	var data [MBRSize]byte

	size, err := d.file.ReadAt(data[:], 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read mbr blob: %w", err)
	}

	if size != MBRSize {
		return nil, fmt.Errorf("%w: mbr read too short", io.ErrUnexpectedEOF)
	}

	return ParseMBR(data[:]), nil
}

func (d *Disk) WriteMBR(mbr *MBR) error {
	var data [MBRSize]byte

	mbr.FillBytes(data[:])

	size, err := d.file.WriteAt(data[:], 0)
	if err != nil {
		return fmt.Errorf("failed to write mbr blob: %w", err)
	}

	if size != MBRSize {
		return fmt.Errorf("%w: mbr write too short", io.ErrUnexpectedEOF)
	}

	return nil
}

func (d *Disk) ReadGPT(lba uint64) (*GPT, error) {
	var data [GPTSize]byte

	size, err := d.file.ReadAt(data[:], int64(lba*BlockSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read gpt blob: %w", err)
	}

	if size != GPTSize {
		return nil, fmt.Errorf("%w: gpt read too short", io.ErrUnexpectedEOF)
	}

	return ParseGPT(data[:])
}

func (d *Disk) WriteGPT(lba uint64, gpt *GPT) error {
	var data [GPTSize]byte

	gpt.FillBytes(data[:])

	size, err := d.file.WriteAt(data[:], int64(lba*BlockSize))
	if err != nil {
		return fmt.Errorf("failed to write gpt blob: %w", err)
	}

	if size != GPTSize {
		return fmt.Errorf("%w: gpt write too short", io.ErrUnexpectedEOF)
	}

	return nil
}

func (d *Disk) ReadGPTPartitions(start uint64, size uint32, count uint32) ([]GPTPartition, uint32, error) {
	return ParseGPTPartitions(d.file, start, size, count)
}

func (d *Disk) WriteGPTPartitions(start uint64, size uint32, parts []GPTPartition) (uint32, error) {
	return WriteGPTPartitions(d.file, start, size, parts)
}
