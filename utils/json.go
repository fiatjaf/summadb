package utils

import "bytes"

func JSONString(s string) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.WriteRune('"')
	for _, b := range s {
		switch b {
		case '\\', '"':
			buffer.WriteRune('\\')
			buffer.WriteRune(b)
			continue
		case '\n':
			buffer.WriteRune('\\')
			buffer.WriteRune('n')
			continue
		case '\r':
			buffer.WriteRune('\\')
			buffer.WriteRune('r')
			continue
		case '\t':
			buffer.WriteRune('\\')
			buffer.WriteRune('t')
			continue
		}
		buffer.WriteRune(b)
	}
	buffer.WriteRune('"')
	return buffer.Bytes()
}
