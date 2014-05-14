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
	outsideCsiCode csiState = iota
	firstCsiCode
	secondeCsiCode
)

type ansiColorWriter struct {
	w        io.Writer
	state    csiState
	paramBuf bytes.Buffer
	textBuf  bytes.Buffer
}

const (
	firstCsiChar   byte = '\x1b'
	secondeCsiChar byte = '['
	separatorChar  byte = ';'
	sgrCode        byte = 'm'
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
	winForegroundColor := wAttributes & (foregroundRed | foregroundGreen | foregroundBlue)
	winBackgroundColor := wAttributes & (backgroundRed | backgroundGreen | backgroundBlue)
	isWinIntensity := (wAttributes & foregroundIntensity) != 0
	csiParam := strings.Split(string(param), string(separatorChar))
	for _, p := range csiParam {
		c, ok := colorMap[p]
		switch {
		case !ok:
			switch p {
			case ansiReset:
				winForegroundColor = foregroundRed | foregroundGreen | foregroundBlue
				winBackgroundColor = 0
				isWinIntensity = false
			case ansiIntensityOn:
				isWinIntensity = true
			case ansiIntensityOff:
				isWinIntensity = false
			default:
				// unknown code
			}
		case c.drawType == foreground:
			winForegroundColor = c.code
		case c.drawType == background:
			winBackgroundColor = c.code
		}
	}
	if isWinIntensity {
		winForegroundColor |= foregroundIntensity
	}
	setConsoleTextAttribute(uintptr(syscall.Stdout), winForegroundColor|winBackgroundColor)
}

func parseEscapeSequence(command byte, param []byte) {
	switch command {
	case sgrCode:
		changeColor(param)
	}
}

func isParameterChar(b byte) bool {
	return ('0' <= b && b <= '9') || b == separatorChar
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
		case outsideCsiCode:
			if ch == firstCsiChar {
				cw.state = firstCsiCode
			} else {
				cw.pushBuffer(ch)
			}
		case firstCsiCode:
			switch ch {
			case firstCsiChar:
				cw.pushBuffer(ch)
			case secondeCsiChar:
				cw.state = secondeCsiCode
			default:
				cw.pushBuffer(firstCsiChar)
				cw.pushBuffer(ch)
				cw.state = outsideCsiCode
			}
		case secondeCsiCode:
			if isParameterChar(ch) {
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
				cw.state = outsideCsiCode
			}
		default:
			cw.pushBuffer(ch)
			cw.state = outsideCsiCode
		}
	}

	nw, err := cw.flushBuffer()
	return r + nw, err
}
