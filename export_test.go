// +build windows

package ansicolor

import "syscall"

var GetConsoleScreenBufferInfo = getConsoleScreenBufferInfo

func ChangeColor(color uint16) {
	setConsoleTextAttribute(uintptr(syscall.Stdout), color)
}

func ResetColor() {
	ChangeColor(uint16(0x0007))
}
