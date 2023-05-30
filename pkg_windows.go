package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.design/x/hotkey"
	"golang.org/x/sys/windows/registry"
)

var (
	ModKeyName = map[string]string{
		"LeftControl":  "Ctrl",
		"RightControl": "Ctrl",
		"LeftAlt":      "Alt",
		"RightAlt":     "Alt",
		"LeftShift":    "Shift",
		"RightShift":   "Shift",
	}

	Modkeys = map[string]int{
		"Ctrl":  0x2,
		"Shift": 0x4,
		"Alt":   0x1,
	}

	Keys = map[string]hotkey.Key{
		"`":  0xC0,
		"/":  0xBF,
		".":  0xBE,
		"-":  0xBD,
		",":  0xBC,
		"+":  0xBB,
		";":  0xBA,
		"[":  0xDB,
		"\\": 0xDC,
		"]":  0xDD,

		"Space": 0x20,
		"0":     0x30,
		"1":     0x31,
		"2":     0x32,
		"3":     0x33,
		"4":     0x34,
		"5":     0x35,
		"6":     0x36,
		"7":     0x37,
		"8":     0x38,
		"9":     0x39,
		"A":     0x41,
		"B":     0x42,
		"C":     0x43,
		"D":     0x44,
		"E":     0x45,
		"F":     0x46,
		"G":     0x47,
		"H":     0x48,
		"I":     0x49,
		"J":     0x4A,
		"K":     0x4B,
		"L":     0x4C,
		"M":     0x4D,
		"N":     0x4E,
		"O":     0x4F,
		"P":     0x50,
		"Q":     0x51,
		"R":     0x52,
		"S":     0x53,
		"T":     0x54,
		"U":     0x55,
		"V":     0x56,
		"W":     0x57,
		"X":     0x58,
		"Y":     0x59,
		"Z":     0x5A,

		"Return": 0x0D,
		"Escape": 0x1B,
		"Delete": 0x2E,
		"Tab":    0x09,

		"Left":  0x25,
		"Right": 0x27,
		"Up  ":  0x26,
		"Down":  0x28,

		"F1":  0x70,
		"F2":  0x71,
		"F3":  0x72,
		"F4":  0x73,
		"F5":  0x74,
		"F6":  0x75,
		"F7":  0x76,
		"F8":  0x77,
		"F9":  0x78,
		"F10": 0x79,
		"F11": 0x7A,
		"F12": 0x7B,
		"F13": 0x7C,
		"F14": 0x7D,
		"F15": 0x7E,
		"F16": 0x7F,
		"F17": 0x80,
		"F18": 0x81,
		"F19": 0x82,
		"F20": 0x83,
	}
)

func RegBase() registry.Key {
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		notification("添加开机自启失败！")
	}

	return k
}

func AutoStart(b bool) {
	exePath, _ := os.Executable()
	_, fn := filepath.Split(exePath)
	k := RegBase()
	defer k.Close()

	if b {
		k.SetStringValue(fn, fmt.Sprintf("%s -autostart", exePath))
	} else {
		k.DeleteValue(fn)
	}
}

func GetAutoStartStatus() bool {
	exePath, _ := os.Executable()
	_, fn := filepath.Split(exePath)

	k := RegBase()
	defer k.Close()
	if _, _, err := k.GetStringValue(fn); err == nil {
		return true
	}

	return false
}
