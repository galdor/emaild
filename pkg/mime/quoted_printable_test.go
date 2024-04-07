package mime

import "testing"

func TestQuotedPrintableEncode(t *testing.T) {
	tests := []struct {
		s  string
		es string
	}{
		{"", ""},
		{"foo", "foo"},
		{"été", "=C3=A9t=C3=A9"},
		{"\r\nfoo\nbar\r\n", "\r\nfoo\r\nbar\r\n"},
		{"foo bar\n baz", "foo bar\r\n baz"},
		{"foo \n", "foo=20\r\n"},
		{"\r\n\rfoo\n\r\nbar", "\r\n=0Dfoo\r\n\r\nbar"},
		{"foo \t \r\n", "foo \t=20\r\n"},

		// 76 characters
		{
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx",
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx",
		},

		// 77 characters
		{
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxy",
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx=\r\ny",
		},

		// 78 characters
		{
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz",
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx=\r\nyz",
		},

		// 2x76 characters
		{
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx\r\n" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx\r\n",
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx\r\n" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx\r\n",
		},

		// 2x78 characters
		{
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz\r\n" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz\r\n",
			"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx=\r\n" +
				"yz\r\n" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwxyz" +
				"abcdefghijklmnopqrstuvwx=\r\n" +
				"yz\r\n",
		},
	}

	for _, test := range tests {
		es := QuotedPrintableEncode(test.s)
		if es != test.es {
			t.Errorf("%q is encoded to %q but should be encoded to %q",
				test.s, es, test.es)
		}
	}
}
