package utils

import "fmt"

func QuoteByte(c byte) string {
	switch {
	case c >= 7 && c <= 13:
		return fmt.Sprintf("%q", c)
	case c < 32:
		return fmt.Sprintf("0x%x", c)
	case c < 127:
		return fmt.Sprintf("%q", c)
	default:
		return fmt.Sprintf("0x%x", c)
	}
}
