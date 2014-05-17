// +build windows

package ansicolor_test

import (
	"bytes"
	"fmt"
	"syscall"
	"testing"

	"github.com/shiena/ansicolor"
	. "github.com/shiena/ansicolor"
)

func TestWritePlanText(t *testing.T) {
	inner := bytes.NewBufferString("")
	w := ansicolor.NewAnsiColorWriter(inner)
	expected := "plain text"
	fmt.Fprintf(w, expected)
	actual := inner.String()
	if actual != expected {
		t.Errorf("Get %v, want %v", actual, expected)
	}
}

type screenNotFoundError struct {
	error
}

func writeAnsiColor(expectedText, colorCode string) (actualText string, actualAttributes uint16, err error) {
	inner := bytes.NewBufferString("")
	w := ansicolor.NewAnsiColorWriter(inner)
	fmt.Fprintf(w, "\x1b[%sm%s", colorCode, expectedText)

	actualText = inner.String()
	screenInfo := GetConsoleScreenBufferInfo(uintptr(syscall.Stdout))
	if screenInfo != nil {
		actualAttributes = screenInfo.WAttributes
	} else {
		err = &screenNotFoundError{}
	}
	return
}

type testParam struct {
	text       string
	attributes uint16
	ansiColor  string
}

func TestWriteAnsiColorText(t *testing.T) {
	screenInfo := GetConsoleScreenBufferInfo(uintptr(syscall.Stdout))
	if screenInfo == nil {
		t.Fatal("Could not get ConsoleScreenBufferInfo")
	}
	defer ChangeColor(screenInfo.WAttributes)

	fgParam := []testParam{
		{"foreground black", uint16(0x0000), "30"},
		{"foreground red", uint16(0x0004), "31"},
		{"foreground green", uint16(0x0002), "32"},
		{"foreground yellow", uint16(0x0006), "33"},
		{"foreground blue", uint16(0x0001), "34"},
		{"foreground magenta", uint16(0x0005), "35"},
		{"foreground cyan", uint16(0x0003), "36"},
		{"foreground white", uint16(0x0007), "37"},
		{"foreground default", uint16(0x0007), "39"},
	}

	bgParam := []testParam{
		{"background black", uint16(0x0007 | 0x0000), "40"},
		{"background red", uint16(0x0007 | 0x0040), "41"},
		{"background green", uint16(0x0007 | 0x0020), "42"},
		{"background yellow", uint16(0x0007 | 0x0060), "43"},
		{"background blue", uint16(0x0007 | 0x0010), "44"},
		{"background magenta", uint16(0x0007 | 0x0050), "45"},
		{"background cyan", uint16(0x0007 | 0x0030), "46"},
		{"background white", uint16(0x0007 | 0x0070), "47"},
		{"background default", uint16(0x0007 | 0x0000), "49"},
	}

	resetParam := []testParam{
		{"all reset", uint16(screenInfo.WAttributes), "0"},
	}

	boldParam := []testParam{
		{"bold on", uint16(0x0007 | 0x0008), "1"},
		{"bold off", uint16(0x0007), "21"},
	}

	mixedParam := []testParam{
		{"both black and bold", uint16(0x0000 | 0x0000 | 0x0008), "30;40;1"},
		{"both red and bold", uint16(0x0004 | 0x0040 | 0x0008), "31;41;1"},
		{"both green and bold", uint16(0x0002 | 0x0020 | 0x0008), "32;42;1"},
		{"both yellow and bold", uint16(0x0006 | 0x0060 | 0x0008), "33;43;1"},
		{"both blue and bold", uint16(0x0001 | 0x0010 | 0x0008), "34;44;1"},
		{"both magenta and bold", uint16(0x0005 | 0x0050 | 0x0008), "35;45;1"},
		{"both cyan and bold", uint16(0x0003 | 0x0030 | 0x0008), "36;46;1"},
		{"both white and bold", uint16(0x0007 | 0x0070 | 0x0008), "37;47;1"},
		{"both default and bold", uint16(0x0007 | 0x0000 | 0x0008), "39;49;1"},
	}

	assertTextAttribute := func(expectedText string, expectedAttributes uint16, ansiColor string) {
		actualText, actualAttributes, err := writeAnsiColor(expectedText, ansiColor)
		if actualText != expectedText {
			t.Errorf("Get %s, want %s", actualText, expectedText)
		}
		if err != nil {
			t.Fatal("Could not get ConsoleScreenBufferInfo")
		}
		if actualAttributes != expectedAttributes {
			t.Errorf("Text: %s, Get %d, want %d", expectedText, actualAttributes, expectedAttributes)
		}
	}

	for _, v := range fgParam {
		ResetColor()
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	for _, v := range bgParam {
		ChangeColor(uint16(0x0070 | 0x0007))
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	for _, v := range resetParam {
		ChangeColor(uint16(0x0000 | 0x0070 | 0x0008))
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	for _, v := range boldParam {
		ResetColor()
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	for _, v := range mixedParam {
		ResetColor()
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}
}
