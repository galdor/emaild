package mime

// RFC 2045 6.8. Base64 Content-Transfer-Encoding

// It would be nice to be able to use the base64 standard package, but it has no
// notion of line length limit.

func Base64Encode(data []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	const maxLineLength = 76

	nbGroups := len(data) / 3
	if len(data)%3 > 0 {
		nbGroups++
	}

	bufSize := nbGroups * 4
	nbEOLs := bufSize / maxLineLength
	if nbEOLs > 0 && bufSize%maxLineLength == 0 {
		nbEOLs--
	}
	bufSize += nbEOLs * 2

	buf := make([]byte, bufSize)

	i := 0
	j := 0

	for i < len(data)/3*3 {
		buf[j+0] = alphabet[(data[i]&0xfc)>>2]
		buf[j+1] = alphabet[((data[i]&0x03)<<4)+((data[i+1]&0xf0)>>4)]
		buf[j+2] = alphabet[((data[i+1]&0x0f)<<2)+((data[i+2]&0xc0)>>6)]
		buf[j+3] = alphabet[data[i+2]&0x3f]

		i += 3
		j += 4

		if j == maxLineLength && i < len(data) {
			buf[j] = '\r'
			buf[j+1] = '\n'
			j += 2
		}
	}

	switch len(data) % 3 {
	case 1:
		buf[j+0] = alphabet[(data[i]&0xfc)>>2]
		buf[j+1] = alphabet[((data[i] & 0x03) << 4)]
		buf[j+2] = '='
		buf[j+3] = '='

	case 2:
		buf[j+0] = alphabet[(data[i]&0xfc)>>2]
		buf[j+1] = alphabet[((data[i]&0x03)<<4)+((data[i+1]&0xf0)>>4)]
		buf[j+2] = alphabet[((data[i+1] & 0x0f) << 2)]
		buf[j+3] = '='
	}

	return string(buf)
}
