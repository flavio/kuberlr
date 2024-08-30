//go:build linux || darwin
// +build linux darwin

package osexec

// Ext is the filename extension of binaries.
const Ext = ""

// TrimExt returns the filename unaltered
// there is no binary extension appended to files on Linux and Mac.
func TrimExt(filename string) string {
	return filename
}
