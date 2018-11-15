// +build windows

package sstr

import (
	"encoding/hex"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	// df9d8cd0-1501-11d1-8c7a-00c04fc297eb
	DefaultCryptoProvider   = "d08c9ddf0115d1118c7a00c04fc297eb"
	CryptProtectUiForbidden = 0x1
)

var (
	modcrypt32           = syscall.NewLazyDLL("crypt32.dll")
	procCryptProtectData = modcrypt32.NewProc("CryptProtectData")
	procUnprotectData    = modcrypt32.NewProc("CryptUnprotectData")
)

func IsProtected(s string) bool {
	return strings.Contains(strings.ToLower(s), DefaultCryptoProvider)
}

func ProtectString(s string) (string, error) {
	b, err := Protect(toUTF16LE(s))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func UnprotectString(s string) (string, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	b, err = Unprotect(b)
	if err != nil {
		return "", err
	}
	return fromUTF16LE(b)
}

func Protect(data []byte) ([]byte, error) {
	inBlob := NewBlob(data)
	var outBlob Blob
	defer outBlob.free()
	r, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(inBlob)),
		0, 0, 0, 0, CryptProtectUiForbidden,
		uintptr(unsafe.Pointer(&outBlob)))
	if r == 0 {
		return nil, err
	}
	return outBlob.bytes(), nil
}

func Unprotect(data []byte) ([]byte, error) {
	inBlob := NewBlob(data)
	var outBlob Blob
	defer outBlob.free()
	r, _, err := procUnprotectData.Call(
		uintptr(unsafe.Pointer(inBlob)),
		0, 0, 0, 0, CryptProtectUiForbidden,
		uintptr(unsafe.Pointer(&outBlob)))
	if r == 0 {
		return nil, err
	}
	return outBlob.bytes(), nil
}
