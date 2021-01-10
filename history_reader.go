package httpproxy

import "io"

func NewHistoryReader(r io.Reader) *HistoryReader {
	return &HistoryReader{
		reader: r,
	}
}

type HistoryReader struct {
	io.Reader
	history []byte
	reader  io.Reader
}

func (hr *HistoryReader) Read(p []byte) (n int, err error) {
	n, err = hr.reader.Read(p)
	if n > 0 {
		hr.history = append(hr.history, p[0:n]...)
	}
	return n, err
}

func (hr *HistoryReader) History() []byte {
	return hr.history
}
