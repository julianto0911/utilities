package utilities

import (
	"bytes"
	"os"
)

func FileStream(filename string) ([]byte, *bytes.Reader, error) {
	v, err := os.ReadFile(filename) //read the content of file
	if err != nil {
		return nil, nil, err
	}

	return v, bytes.NewReader(v), nil
}
