//go:build darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.design/x/hotkey"
)

var (
	ModKeyName = map[string]string{
		"LeftControl":  "Ctrl",
		"RightControl": "Ctrl",
		"LeftAlt":      "Option",
		"RightAlt":     "Option",
		"LeftShift":    "Shift",
		"RightShift":   "Shift",
		"LeftSuper":    "Cmd",
		"RightSuper":   "Cmd",
	}
	Modkeys = map[string]int{
		"Ctrl":   0x1000,
		"Shift":  0x200,
		"Option": 0x800,
		"Cmd":    0x100,
	}
	Keys = map[string]hotkey.Key{
		"`":  0x32,
		"/":  0x2C,
		".":  0x2F,
		"-":  0x1B,
		",":  0x2B,
		"+":  0x18,
		";":  0x29,
		"[":  0x21,
		"\\": 0x2A,
		"]":  0x1E,

		"Space": 49,
		"1":     18,
		"2":     19,
		"3":     20,
		"4":     21,
		"5":     23,
		"6":     22,
		"7":     26,
		"8":     28,
		"9":     25,
		"0":     29,
		"A":     0,
		"B":     11,
		"C":     8,
		"D":     2,
		"E":     14,
		"F":     3,
		"G":     5,
		"H":     4,
		"I":     34,
		"J":     38,
		"K":     40,
		"L":     37,
		"M":     46,
		"N":     45,
		"O":     31,
		"P":     35,
		"Q":     12,
		"R":     15,
		"S":     1,
		"T":     17,
		"U":     32,
		"V":     9,
		"W":     13,
		"X":     7,
		"Y":     16,
		"Z":     6,

		"Return": 0x24,
		"Escape": 0x35,
		"Delete": 0x33,
		"Tab   ": 0x30,

		"Left ": 0x7B,
		"Right": 0x7C,
		"Up   ": 0x7E,
		"Down ": 0x7D,

		"F1 ": 0x7A,
		"F2 ": 0x78,
		"F3 ": 0x63,
		"F4 ": 0x76,
		"F5 ": 0x60,
		"F6 ": 0x61,
		"F7 ": 0x62,
		"F8 ": 0x64,
		"F9 ": 0x65,
		"F10": 0x6D,
		"F11": 0x67,
		"F12": 0x6F,
		"F13": 0x69,
		"F14": 0x6B,
		"F15": 0x71,
		"F16": 0x6A,
		"F17": 0x40,
		"F18": 0x4F,
		"F19": 0x50,
		"F20": 0x5A,
	}
)

func AutoStart(b bool) {
	home, _ := os.UserHomeDir()
	path := fmt.Sprintf("%s/Library/LaunchAgents/autoTyper.plist", home)

	if b {
		exePath, _ := os.Executable()
		p, _ := filepath.Split(exePath)
		os.WriteFile(path, []byte(fmt.Sprintf(PlistContent, exePath, p)), os.ModePerm)
	} else {
		os.Remove(path)
	}
}

func GetAutoStartStatus() bool {
	home, _ := os.UserHomeDir()
	path := fmt.Sprintf("%s/Library/LaunchAgents/autoTyper.plist", home)
	_, err := os.Stat(path)

	return err == nil
}

var PlistContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.tw.autoTyper</string>
  <key>RunAtLoad</key>
  <true/>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>-autostart</string>
  </array>
  <key>WorkingDirectory</key>
  <string>%s</string>
</dict>
</plist>
`
