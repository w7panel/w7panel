package webdav2

import (
	"bytes"
	"io"
	"sync"
)

type SeekableBuffer struct {
	mu sync.RWMutex
	bytes.Buffer
	offset int64
	length int64
}

func NewSeekableBuffer() *SeekableBuffer {
	return &SeekableBuffer{}
}

func (s *SeekableBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, err = s.Buffer.Write(p)
	s.length += int64(n)
	return
}

func (s *SeekableBuffer) Read(p []byte) (n int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.offset >= s.length {
		return 0, io.EOF
	}

	data := s.Buffer.Bytes()[s.offset:]
	if len(data) < len(p) {
		n = len(data)
	} else {
		n = len(p)
	}

	copy(p, data[:n])
	s.offset += int64(n)
	return n, nil
}

func (s *SeekableBuffer) Seek(offset int64, whence int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = s.offset + offset
	case io.SeekEnd:
		newOffset = s.length + offset
	default:
		return 0, io.ErrUnexpectedEOF
	}

	if newOffset < 0 {
		return 0, io.ErrUnexpectedEOF
	}

	s.offset = newOffset
	return s.offset, nil
}
