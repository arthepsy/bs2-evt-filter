// +build windows

package sstr

import (
	"bytes"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

func toUTF16LE(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, 2*len(u))
	for index, value := range u {
		b[index*2] = byte(value)
		b[index*2+1] = byte(value >> 8)
	}
	return b
}

func fromUTF16LE(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}
	u16s := make([]uint16, 1)
	b8buf := make([]byte, 4)
	ret := &bytes.Buffer{}
	for i := 0; i < len(b); i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}
	return ret.String(), nil
}
