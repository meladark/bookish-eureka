package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func printProgressBar(copied int64, total int64) {
	const width = 50
	percent := float64(copied) / float64(total) * 100
	filled := int(float64(width) * percent / 100)
	empty := width - filled

	fmt.Printf("\r[%s%s] %.2f%%",
		strings.Repeat("=", filled),
		strings.Repeat(" ", empty),
		percent,
	)
}

func Copy(fromPath string, toPath string, offset int64, limit int64) error {
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer fromFile.Close()
	info, err := fromFile.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	fileSize := info.Size()
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}
	_, err = fromFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	toFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer toFile.Close()
	bytesLeft := fileSize - offset
	if limit == 0 || limit > bytesLeft {
		limit = bytesLeft
	}
	const bufSize = 4096
	buf := make([]byte, bufSize)
	var copied int64
	printProgressBar(0, limit)
	for copied < limit {
		bytesToRead := bufSize
		if remaining := limit - copied; remaining < int64(bufSize) {
			bytesToRead = int(remaining)
		}
		n, err := fromFile.Read(buf[:bytesToRead])
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if n > 0 {
			_, err := toFile.Write(buf[:n])
			if err != nil {
				return err
			}
			copied += int64(n)
			printProgressBar(copied, limit)
		}
	}
	return nil
}
