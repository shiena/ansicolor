package ansicolor

import "io"

func NewAnsiColorWriter(w io.Writer) *ansiColorWriter {
	return &ansiColorWriter{w: w}
}
