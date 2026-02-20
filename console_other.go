//go:build !windows

package main

// HideConsole is a no-op on non-Windows platforms
func HideConsole() {}

// ShowConsoleWindow is a no-op on non-Windows platforms
func ShowConsoleWindow() {}
