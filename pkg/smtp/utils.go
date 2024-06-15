package smtp

import (
	"fmt"

	"github.com/galdor/emaild/pkg/imf"
)

func ValidateDomain(data []byte) (string, error) {
	decoder := imf.NewDataDecoder(data)

	domain, err := decoder.ReadDomain()
	if err != nil {
		return "", err
	}

	if !decoder.Empty() {
		return "", fmt.Errorf("invalid trailing data")
	}

	return string(*domain), nil
}
