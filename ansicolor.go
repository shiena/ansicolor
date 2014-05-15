// Package ansicolor provides color console in Windows as ANSICON.
package ansicolor

import "io"

// NewAnsiColorWriter creates and initializes a new ansiColorWriter
// using io.Writer w as its initial contents.
// In the console of Windows, which change the foreground and background
// colors of the text by the escape sequence.
// In the console of other systems, which writes to w all text.
func NewAnsiColorWriter(w io.Writer) *ansiColorWriter {
	return &ansiColorWriter{w: w}
}
