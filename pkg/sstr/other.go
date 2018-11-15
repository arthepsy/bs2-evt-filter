// +build !windows

package sstr

func IsProtected(s string) bool {
	return false
}

func ProtectString(s string) (string, error) {
	return s, nil
}

func UnprotectString(s string) (string, error) {
	return s, nil
}

func Protect(data []byte) ([]byte, error) {
	return data, nil
}

func Unprotect(data []byte) ([]byte, error) {
	return data, nil
}
