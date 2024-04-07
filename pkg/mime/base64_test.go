package mime

import (
	"testing"
)

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		s  string
		es string
	}{
		{"", ""},
		{"f", "Zg=="},
		{"fo", "Zm8="},
		{"foo", "Zm9v"},
		{"foob", "Zm9vYg=="},
		{"fooba", "Zm9vYmE="},
		{"foobar", "Zm9vYmFy"},

		// 57 bytes (76 encoded characters)
		{
			"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcde",
			"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXphYmNkZWZnaGlqa2xtbm9wcXJzdHV2d3h5emFiY2Rl",
		},

		// 58 bytes (80 encoded characters)
		{
			"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdef",
			"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXphYmNkZWZnaGlqa2xtbm9wcX" +
				"JzdHV2d3h5emFiY2Rl\r\nZg==",
		},
	}

	for _, test := range tests {
		es := Base64Encode([]byte(test.s))
		if es != test.es {
			t.Errorf("%q is encoded to %q but should be encoded to %q",
				test.s, es, test.es)
		}
	}
}
