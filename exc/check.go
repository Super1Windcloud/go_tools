package main

import (
	"bytes"
	"os"
)

// 常见压缩格式的 magic number 映射
var magicNumbers = map[string][]byte{
	"zip": {0x50, 0x4B, 0x03, 0x04},
	"gz":  {0x1F, 0x8B},
	"7z":  {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
	"rar": {0x52, 0x61, 0x72, 0x21},
	"xz":  {0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00},
}

func IsCompressedByMagic(path string) (bool, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, "", err
	}
	defer file.Close()


	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		return false, "", err
	}

	for typ, magic := range magicNumbers {
		if bytes.HasPrefix(header, magic) {
			return true, typ, nil
		}
	}

	return false, "", nil
}
