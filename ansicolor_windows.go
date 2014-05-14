// +build windows

package ansicolor

import (
	"bytes"
	"io"
	"strings"
	"syscall"
	"unsafe"
)

type csiState int

const (
	TEXT csiState = iota
	ESC1
	ESC2
)

type ansiColorWriter struct {
	w        io.Writer
	state    csiState
	paramBuf bytes.Buffer
	textBuf  bytes.Buffer
}

const (
	CSI1  byte = '\x1b'
	CSI2  byte = '['
	SEP   byte = ';'
	COLOR byte = 'm'
)

const (
	foregroundBlue      = uint16(0x0001)
	foregroundGreen     = uint16(0x0002)
	foregroundRed       = uint16(0x0004)
	foregroundIntensity = uint16(0x0008)
	backgroundBlue      = uint16(0x0010)
	backgroundGreen     = uint16(0x0020)
	backgroundRed       = uint16(0x0040)
	backgroundIntensity = uint16(0x0080)

	foregroundMask = foregroundBlue | foregroundGreen | foregroundRed | foregroundIntensity
	backgroundMask = backgroundBlue | backgroundGreen | backgroundRed | backgroundIntensity
)

const (
	ansiReset        = "0"
	ansiIntensityOn  = "1"
	ansiIntensityOff = "22"

	ansiForegroundBlack   = "30"
	ansiForegroundRed     = "31"
	ansiForegroundGreen   = "32"
	ansiForegroundYellow  = "33"
	ansiForegroundBlue    = "34"
	ansiForegroundMagenta = "35"
	ansiForegroundCyan    = "36"
	ansiForegroundWhite   = "37"
	ansiForegroundDefault = "39"

	ansiBackgroundBlack   = "40"
	ansiBackgroundRed     = "41"
	ansiBackgroundGreen   = "42"
	ansiBackgroundYellow  = "43"
	ansiBackgroundBlue    = "44"
	ansiBackgroundMagenta = "45"
	ansiBackgroundCyan    = "46"
	ansiBackgroundWhite   = "47"
	ansiBackgroundDefault = "49"
)

type drawType int

const (
	foreground drawType = iota
	background
)

type winColor struct {
	code     uint16
	drawType drawType
}

var colorMap = map[string]winColor{
	ansiForegroundBlack:   {0, foreground},
	ansiForegroundRed:     {foregroundRed, foreground},
	ansiForegroundGreen:   {foregroundGreen, foreground},
	ansiForegroundYellow:  {foregroundRed | foregroundGreen, foreground},
	ansiForegroundBlue:    {foregroundBlue, foreground},
	ansiForegroundMagenta: {foregroundRed | foregroundBlue, foreground},
	ansiForegroundCyan:    {foregroundGreen | foregroundBlue, foreground},
	ansiForegroundWhite:   {foregroundRed | foregroundGreen | foregroundBlue, foreground},
	ansiForegroundDefault: {foregroundRed | foregroundGreen | foregroundBlue, foreground},

	ansiBackgroundBlack:   {0, background},
	ansiBackgroundRed:     {backgroundRed, background},
	ansiBackgroundGreen:   {backgroundGreen, background},
	ansiBackgroundYellow:  {backgroundRed | backgroundGreen, background},
	ansiBackgroundBlue:    {backgroundBlue, background},
	ansiBackgroundMagenta: {backgroundRed | backgroundBlue, background},
	ansiBackgroundCyan:    {backgroundGreen | backgroundBlue, background},
	ansiBackgroundWhite:   {backgroundRed | backgroundGreen | backgroundBlue, background},
	ansiBackgroundDefault: {0, background},
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTextAttribute    = kernel32.NewProc("SetConsoleTextAttribute")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct {
	X, Y int16
}

type smallRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	DwSize              coord
	DwCursorPosition    coord
	WAttributes         uint16
	SrWindow            smallRect
	DwMaximumWindowSize coord
}

func getConsoleScreenBufferInfo(hConsoleOutput uintptr) *consoleScreenBufferInfo {
	var csbi consoleScreenBufferInfo
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		hConsoleOutput,
		uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		return nil
	}
	return &csbi
}

func setConsoleTextAttribute(hConsoleOutput uintptr, wAttributes uint16) bool {
	ret, _, _ := procSetConsoleTextAttribute.Call(
		hConsoleOutput,
		uintptr(wAttributes))
	return ret != 0
}

func changeColor(param []byte) {
	screenInfo := getConsoleScreenBufferInfo(uintptr(syscall.Stdout))
	if screenInfo == nil {
		return
	}

	wAttributes := screenInfo.WAttributes
	winForeColor := wAttributes & (foregroundRed | foregroundGreen | foregroundBlue)
	winBackColor := wAttributes & (backgroundRed | backgroundGreen | backgroundBlue)
	winIntensity := (wAttributes & foregroundIntensity) != 0
	paramLine := strings.Split(string(param), string(SEP))
	for _, p := range paramLine {
		c, ok := colorMap[p]
		switch {
		case !ok:
			switch p {
			case ansiReset:
				winForeColor = foregroundRed | foregroundGreen | foregroundBlue
				winBackColor = 0
				winIntensity = false
			case ansiIntensityOn:
				winIntensity = true
			case ansiIntensityOff:
				winIntensity = false
			default:
				// unknown code
			}
		case c.drawType == foreground:
			winForeColor = c.code
		case c.drawType == background:
			winBackColor = c.code
		}
	}
	if winIntensity {
		winForeColor |= foregroundIntensity
	}
	setConsoleTextAttribute(uintptr(syscall.Stdout), winForeColor|winBackColor)
}

func parseEscapeSequence(command byte, param []byte) {
	switch command {
	case COLOR:
		changeColor(param)
	}
}

func isParam(b byte) bool {
	return ('0' <= b && b <= '9') || b == SEP
}

func (cw *ansiColorWriter) pushBuffer(ch byte) {
	cw.textBuf.WriteByte(ch)
}

func (cw *ansiColorWriter) flushBuffer() (int, error) {
	text := cw.textBuf.Bytes()
	cw.textBuf.Reset()
	return cw.w.Write(text)
}

func (cw *ansiColorWriter) Write(p []byte) (int, error) {
	r := 0
	for _, ch := range p {
		switch cw.state {
		case TEXT:
			if ch == CSI1 {
				cw.state = ESC1
			} else {
				cw.pushBuffer(ch)
			}
		case ESC1:
			switch ch {
			case CSI1:
				cw.pushBuffer(ch)
			case CSI2:
				cw.state = ESC2
			default:
				cw.pushBuffer(CSI1)
				cw.pushBuffer(ch)
				cw.state = TEXT
			}
		case ESC2:
			if isParam(ch) {
				cw.paramBuf.WriteByte(ch)
			} else {
				nw, err := cw.flushBuffer()
				r += nw
				if err != nil {
					return r, err
				}
				param := cw.paramBuf.Bytes()
				cw.paramBuf.Reset()
				parseEscapeSequence(ch, param)
				cw.state = TEXT
			}
		default:
			cw.pushBuffer(ch)
			cw.state = TEXT
		}
	}

	nw, err := cw.flushBuffer()
	return r + nw, err
}
