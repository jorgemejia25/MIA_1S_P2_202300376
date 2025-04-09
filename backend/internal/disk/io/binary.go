package io

import (
	"bytes"
	"encoding/binary"
	"os"
)

// ReadStructAt lee una estructura desde un archivo en una posición específica
func ReadStructAt(file *os.File, data interface{}, pos int64) error {
	buf := make([]byte, binary.Size(data))
	if _, err := file.ReadAt(buf, pos); err != nil {
		return err
	}
	return binary.Read(bytes.NewReader(buf), binary.LittleEndian, data)
}

// WriteStructAt escribe una estructura en un archivo en una posición específica
func WriteStructAt(file *os.File, data interface{}, pos int64) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return err
	}
	_, err := file.WriteAt(buf.Bytes(), pos)
	return err
}
