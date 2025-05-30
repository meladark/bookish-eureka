package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
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
		bar := pb.Full.Start64(limit)
		defer bar.Finish()
		if n > 0 {
			_, err := toFile.Write(buf[:n])
			if err != nil {
				return err
			}
			copied += int64(n)
			bar.Add(n)
		}
	}
	return nil
}
