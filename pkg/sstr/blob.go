// +build windows

package sstr

import (
	"syscall"
	"unsafe"
)

type Blob struct {
	cbData uint32
	pbData *byte
}

var (
	modkernel32   = syscall.NewLazyDLL("kernel32.dll")
	procLocalFree = modkernel32.NewProc("LocalFree")
)

func NewBlob(d []byte) *Blob {
	if len(d) == 0 {
		return &Blob{}
	}
	return &Blob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *Blob) bytes() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func (b *Blob) free() {
	procLocalFree.Call(uintptr(unsafe.Pointer(b.pbData)))
}
