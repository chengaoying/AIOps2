package wal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"aiops2/collector/internal/model"
)

const (
	Magic          = "WAL0"
	Version        = 1
	MaxFileSize    = 1 << 30 // 1GB
	FilePrefix     = "WAL-"
	FileExt        = ".wal"
	RecordInsert   = 1
	RecordUpdate   = 2
	RecordDelete   = 3
)

type Record struct {
	Type      byte
	Timestamp int64
	Platform  string
	JobMeta   *model.JobMeta
	Checksum  uint32
}

type WAL struct {
	dir        string
	maxFileSize int64
	mu         sync.Mutex
	currentFile *os.File
	currentSize int64
	seq        int
}

func New(dir string, maxFileSize int64) (*WAL, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create wal dir: %w", err)
	}

	w := &WAL{
		dir:        dir,
		maxFileSize: maxFileSize,
	}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *WAL) Write(job *model.JobMeta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	rec := Record{
		Type:      RecordInsert,
		Timestamp: time.Now().UnixMilli(),
		Platform:  job.Platform,
		JobMeta:   job,
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	checksum := crc32.ChecksumIEEE(data)
	rec.Checksum = checksum

	recData, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal record with checksum: %w", err)
	}

	length := int32(len(recData))
	if err := binary.Write(w.currentFile, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("write length: %w", err)
	}

	if _, err := w.currentFile.Write(recData); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	w.currentSize += int64(length + 4)

	if w.currentSize >= w.maxFileSize {
		return w.rotate()
	}
	return nil
}

func (w *WAL) rotate() error {
	if w.currentFile != nil {
		w.currentFile.Close()
	}

	w.seq++
	filename := filepath.Join(w.dir, fmt.Sprintf("%s%d-%d%s", FilePrefix, time.Now().UnixMilli(), w.seq, FileExt))

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create wal file: %w", err)
	}

	header := fmt.Sprintf("%s:%d\n", Magic, Version)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	w.currentFile = f
	w.currentSize = int64(len(header))
	return nil
}

func (w *WAL) ReadAll() ([]*model.JobMeta, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	files, err := w.getWALFiles()
	if err != nil {
		return nil, err
	}

	var jobs []*model.JobMeta
	for _, fname := range files {
		f, err := os.Open(filepath.Join(w.dir, fname))
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)

		if scanner.Scan() {
			_ = scanner.Text()
		}

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) < 4 {
				continue
			}

			if len(line) < 4 {
			continue
		}

		lengthBuf := line[:4]
		length := int32(binary.LittleEndian.Uint32(lengthBuf))

		if len(line) < 4+int(length) {
			continue
		}

		var rec Record
		if err := json.Unmarshal(line[4:4+int(length)], &rec); err != nil {
			continue
		}

			if rec.JobMeta != nil {
				jobs = append(jobs, rec.JobMeta)
			}
		}
		f.Close()
	}

	return jobs, nil
}

func (w *WAL) getWALFiles() ([]string, error) {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), FilePrefix) && strings.HasSuffix(e.Name(), FileExt) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

func (w *WAL) ReadAll() ([]*model.JobMeta, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	files, err := w.getWALFiles()
	if err != nil {
		return nil, err
	}

	var jobs []*model.JobMeta
	for _, fname := range files {
		f, err := os.Open(filepath.Join(w.dir, fname))
		if err != nil {
			continue
		}

		br := bufio.NewReader(f)

		line, err := br.ReadBytes('\n')
		if err != nil && err != io.EOF {
			f.Close()
			continue
		}
		_ = string(line)

		for {
			var length int32
			if err := binary.Read(br, binary.LittleEndian, &length); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			data := make([]byte, length)
			if _, err := io.ReadFull(br, data); err != nil {
				continue
			}

			var rec Record
			if err := json.Unmarshal(data, &rec); err != nil {
				continue
			}

			if rec.JobMeta != nil {
				jobs = append(jobs, rec.JobMeta)
			}
		}
		f.Close()
	}

	return jobs, nil
}
