package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type WALRecord struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type WAL struct {
	file   *os.File
	writer *bufio.Writer
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (w *WAL) Append(rec WALRecord) error {
	bytes, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(append(bytes, '\n'))
	if err != nil {
		return err
	}
	return w.writer.Flush()
}

func (w *WAL) ReadAll() ([]WALRecord, error) {
	_, err := w.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	var records []WALRecord
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		var rec WALRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			fmt.Println("Skipping invalid WAL entry:", err)
			continue
		}
		records = append(records, rec)
	}
	return records, scanner.Err()
}
