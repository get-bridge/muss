package proc

import (
	"io"
)

// Pipe has a reader and writer with getters and setters.
type Pipe struct {
	reader io.Reader
	writer io.Writer
}

// Reader returns the reader.
func (p *Pipe) Reader() io.Reader {
	return p.reader
}

// Writer returns the writer.
func (p *Pipe) Writer() io.Writer {
	return p.writer
}

// SetReader sets the reader.
func (p *Pipe) SetReader(r io.Reader) {
	p.reader = r
}

// SetWriter sets the writer.
func (p *Pipe) SetWriter(w io.Writer) {
	p.writer = w
}
