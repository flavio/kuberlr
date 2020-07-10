// +build windows

package osexec

import "strings"

// Ext is the filename extension of binaries
const Ext = ".exe"

// TrimExt returns the filename with the extension removed
func TrimExt(filename string) string {
	return strings.TrimSuffix(filename, Ext)
}
