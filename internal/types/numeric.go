package types

import (
	"strings"
)

type Numeric string

func (n Numeric) MarshalJSON() ([]byte, error) {
	splitted := strings.Split(string(n), ".")

	if len(splitted) == 1 {
		// no decimal point to preserve, can decode as number
		return []byte(n), nil
	}

	// TODO: this is rather arbitrary, find the actual break-off limit for floating point problems as json numbers
	if len(splitted[1]) < 5 {
		return []byte(n), nil
	}

	return []byte("\"" + n + "\""), nil
}
