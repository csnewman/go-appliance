package disk

import (
	"fmt"
	"io"
	"os"
)

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
