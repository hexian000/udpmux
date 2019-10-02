package udpmux

import (
	"io"
	"os"
	"time"
)

var (
	_ = io.WriteCloser((*RotateWriter)(nil))
)

type RotateWriter struct {
	namePattern string
	header      []byte
	f           *os.File
}

func (r *RotateWriter) newFile(name string) error {
	newFile := false
	if _, err := os.Stat(name); err != nil {
		newFile = true
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if newFile {
		if _, err = f.Write(r.header); err != nil {
			_ = f.Close()
			return err
		}
	}
	if r.f != nil {
		_ = r.f.Close()
	}
	r.f = f
	return nil
}

func CreateRotateWriteCloser(namePattern string, header []byte) *RotateWriter {
	return &RotateWriter{namePattern: namePattern, header: header}
}

func (r *RotateWriter) Write(p []byte) (n int, err error) {
	name := time.Now().Format(r.namePattern)
	if r.f == nil || name != r.f.Name() {
		if err = r.newFile(name); err != nil {
			return
		}
	}
	return r.f.Write(p)
}

func (r *RotateWriter) Close() error {
	return r.f.Close()
}
