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
	foreground_blue      = uint16(0x0001)
	foreground_green     = uint16(0x0002)
	foreground_red       = uint16(0x0004)
	foreground_intensity = uint16(0x0008)
	background_blue      = uint16(0x0010)
	background_green     = uint16(0x0020)
	background_red       = uint16(0x0040)
	background_intensity = uint16(0x0080)

	foreground_mask = foreground_blue | foreground_green | foreground_red | foreground_intensity
	background_mask = background_blue | background_green | background_red | background_intensity
)

const (
	ansi_reset         = "0"
	ansi_intensity_on  = "1"
	ansi_intensity_off = "22"

	ansi_foreground_black   = "30"
	ansi_foreground_red     = "31"
	ansi_foreground_green   = "32"
	ansi_foreground_yellow  = "33"
	ansi_foreground_blue    = "34"
	ansi_foreground_magenta = "35"
	ansi_foreground_cyan    = "36"
	ansi_foreground_white   = "37"
	ansi_foreground_default = "39"

	ansi_background_black   = "40"
	ansi_background_red     = "41"
	ansi_background_green   = "42"
	ansi_background_yellow  = "43"
	ansi_background_blue    = "44"
	ansi_background_magenta = "45"
	ansi_background_cyan    = "46"
	ansi_background_white   = "47"
	ansi_background_default = "49"
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
	ansi_foreground_black:   {0, foreground},
	ansi_foreground_red:     {foreground_red, foreground},
	ansi_foreground_green:   {foreground_green, foreground},
	ansi_foreground_yellow:  {foreground_red | foreground_green, foreground},
	ansi_foreground_blue:    {foreground_blue, foreground},
	ansi_foreground_magenta: {foreground_red | foreground_blue, foreground},
	ansi_foreground_cyan:    {foreground_green | foreground_blue, foreground},
	ansi_foreground_white:   {foreground_red | foreground_green | foreground_blue, foreground},
	ansi_foreground_default: {foreground_red | foreground_green | foreground_blue, foreground},

	ansi_background_black:   {0, background},
	ansi_background_red:     {background_red, background},
	ansi_background_green:   {background_green, background},
	ansi_background_yellow:  {background_red | background_green, background},
	ansi_background_blue:    {background_blue, background},
	ansi_background_magenta: {background_red | background_blue, background},
	ansi_background_cyan:    {background_green | background_blue, background},
	ansi_background_white:   {background_red | background_green | background_blue, background},
	ansi_background_default: {0, background},
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTextAttribute    = kernel32.NewProc("SetConsoleTextAttribute")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct {
	X, Y int16
}

type small_rect struct {
	Left, Top, Right, Bottom int16
}

type console_screen_buffer_info struct {
	DwSize              coord
	DwCursorPosition    coord
	WAttributes         uint16
	SrWindow            small_rect
	DwMaximumWindowSize coord
}

func getConsoleScreenBufferInfo(hConsoleOutput uintptr) *console_screen_buffer_info {
	var csbi console_screen_buffer_info
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
	winForeColor := wAttributes & (foreground_red | foreground_green | foreground_blue)
	winBackColor := wAttributes & (background_red | background_green | background_blue)
	winIntensity := (wAttributes & foreground_intensity) != 0
	param_line := strings.Split(string(param), string(SEP))
	for _, p := range param_line {
		c, ok := colorMap[p]
		switch {
		case !ok:
			switch p {
			case ansi_reset:
				winForeColor = foreground_red | foreground_green | foreground_blue
				winBackColor = 0
				winIntensity = false
			case ansi_intensity_on:
				winIntensity = true
			case ansi_intensity_off:
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
		winForeColor |= foreground_intensity
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
