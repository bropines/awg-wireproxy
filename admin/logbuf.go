package admin

import (
	"bytes"
	"sync"
)

// LogLine is a single captured log line.
type LogLine struct {
	Seq  uint64 `json:"seq"`
	Text string `json:"text"`
}

// LogBuffer is an io.Writer that keeps the last lines written to it.
type LogBuffer struct {
	mu      sync.Mutex
	max     int
	lines   []LogLine
	partial []byte
	seq     uint64
}

// NewLogBuffer creates a buffer keeping at most max lines.
func NewLogBuffer(max int) *LogBuffer {
	return &LogBuffer{max: max}
}

// Write implements io.Writer. Input is split on newlines; an unterminated
// tail is kept until the rest of the line arrives.
func (b *LogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	data := append(b.partial, p...)
	b.partial = nil
	for {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		b.add(string(bytes.TrimRight(data[:i], "\r")))
		data = data[i+1:]
	}
	if len(data) > 0 {
		b.partial = append([]byte(nil), data...)
	}
	return len(p), nil
}

func (b *LogBuffer) add(text string) {
	b.seq++
	b.lines = append(b.lines, LogLine{Seq: b.seq, Text: text})
	if len(b.lines) > b.max {
		b.lines = append([]LogLine(nil), b.lines[len(b.lines)-b.max:]...)
	}
}

// Since returns the lines with a sequence number greater than after, and the
// sequence number to pass on the next call.
func (b *LogBuffer) Since(after uint64) ([]LogLine, uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]LogLine, 0)
	for _, l := range b.lines {
		if l.Seq > after {
			out = append(out, l)
		}
	}
	return out, b.seq
}
