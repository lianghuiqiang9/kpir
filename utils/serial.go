package utils

import (
	"encoding/gob"
	"io"
	"os"
	"unsafe"
)

// if exist, return ture
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// save
func SaveToFile(data interface{}, filename string) (int64, error) {
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return 0, err
	}

	file.Sync()

	info, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func LoadFromFile(data interface{}, filename string) (int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	fileSize := info.Size()

	if fileSize == 0 {
		return 0, nil
	}

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return fileSize, err
	}

	return fileSize, nil
}

// SaveUint64Slice
func SaveUint64Slice(data []uint64, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	ptr := unsafe.SliceData(data)
	byteLen := len(data) * 8
	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), byteLen)

	_, err = f.Write(byteSlice)
	return err
}

func LoadUint64Slice(data []uint64, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	ptr := unsafe.SliceData(data)

	byteLen := len(data) * 8
	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), byteLen)

	_, err = io.ReadFull(f, byteSlice)
	return err
}
