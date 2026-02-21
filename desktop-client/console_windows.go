//go:build windows

package main

import (
	"syscall"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	user32           = syscall.NewLazyDLL("user32.dll")
	showWindow       = user32.NewProc("ShowWindow")
)

const swHide = 0

// HideConsole hides the console window on Windows
func HideConsole() {
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		showWindow.Call(hwnd, swHide)
	}
}

// ShowConsoleWindow shows the console window on Windows
func ShowConsoleWindow() {
	// Console is shown by default, nothing to do
}
