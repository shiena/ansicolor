// Copyright 2014 shiena Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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
		t.Errorf("Get 0x%04x, want 0x%04x", actual, expected)
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
	defaultFgColor := screenInfo.WAttributes & uint16(0x0007)
	defaultBgColor := screenInfo.WAttributes & uint16(0x0070)

	fgParam := []testParam{
		{"foreground black  ", uint16(0x0000 | 0x0000), "30"},
		{"foreground red    ", uint16(0x0004 | 0x0000), "31"},
		{"foreground green  ", uint16(0x0002 | 0x0000), "32"},
		{"foreground yellow ", uint16(0x0006 | 0x0000), "33"},
		{"foreground blue   ", uint16(0x0001 | 0x0000), "34"},
		{"foreground magenta", uint16(0x0005 | 0x0000), "35"},
		{"foreground cyan   ", uint16(0x0003 | 0x0000), "36"},
		{"foreground white  ", uint16(0x0007 | 0x0000), "37"},
		{"foreground default", defaultFgColor | 0x0000, "39"},
	}

	bgParam := []testParam{
		{"background black  ", uint16(0x0007 | 0x0000), "40"},
		{"background red    ", uint16(0x0007 | 0x0040), "41"},
		{"background green  ", uint16(0x0007 | 0x0020), "42"},
		{"background yellow ", uint16(0x0007 | 0x0060), "43"},
		{"background blue   ", uint16(0x0007 | 0x0010), "44"},
		{"background magenta", uint16(0x0007 | 0x0050), "45"},
		{"background cyan   ", uint16(0x0007 | 0x0030), "46"},
		{"background white  ", uint16(0x0007 | 0x0070), "47"},
		{"background default", uint16(0x0007) | defaultBgColor, "49"},
	}

	resetParam := []testParam{
		{"all reset", defaultFgColor | defaultBgColor, "0"},
		{"all reset", defaultFgColor | defaultBgColor, ""},
	}

	boldParam := []testParam{
		{"bold on", uint16(0x0007 | 0x0008), "1"},
		{"bold off", uint16(0x0007), "21"},
	}

	underscoreParam := []testParam{
		{"underscore on", uint16(0x0007 | 0x8000), "4"},
		{"underscore off", uint16(0x0007), "24"},
	}

	blinkParam := []testParam{
		{"blink on", uint16(0x0007 | 0x0080), "5"},
		{"blink off", uint16(0x0007), "25"},
	}

	mixedParam := []testParam{
		{"both black,   bold, underline, blink", uint16(0x0000 | 0x0000 | 0x0008 | 0x8000 | 0x0080), "30;40;1;4;5"},
		{"both red,     bold, underline, blink", uint16(0x0004 | 0x0040 | 0x0008 | 0x8000 | 0x0080), "31;41;1;4;5"},
		{"both green,   bold, underline, blink", uint16(0x0002 | 0x0020 | 0x0008 | 0x8000 | 0x0080), "32;42;1;4;5"},
		{"both yellow,  bold, underline, blink", uint16(0x0006 | 0x0060 | 0x0008 | 0x8000 | 0x0080), "33;43;1;4;5"},
		{"both blue,    bold, underline, blink", uint16(0x0001 | 0x0010 | 0x0008 | 0x8000 | 0x0080), "34;44;1;4;5"},
		{"both magenta, bold, underline, blink", uint16(0x0005 | 0x0050 | 0x0008 | 0x8000 | 0x0080), "35;45;1;4;5"},
		{"both cyan,    bold, underline, blink", uint16(0x0003 | 0x0030 | 0x0008 | 0x8000 | 0x0080), "36;46;1;4;5"},
		{"both white,   bold, underline, blink", uint16(0x0007 | 0x0070 | 0x0008 | 0x8000 | 0x0080), "37;47;1;4;5"},
		{"both default, bold, underline, blink", uint16(defaultFgColor | defaultBgColor | 0x0008 | 0x8000 | 0x0080), "39;49;1;4;5"},
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
			t.Errorf("Text: %s, Get 0x%04x, want 0x%04x", expectedText, actualAttributes, expectedAttributes)
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

	ResetColor()
	for _, v := range boldParam {
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	ResetColor()
	for _, v := range underscoreParam {
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	ResetColor()
	for _, v := range blinkParam {
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}

	for _, v := range mixedParam {
		ResetColor()
		assertTextAttribute(v.text, v.attributes, v.ansiColor)
	}
}
