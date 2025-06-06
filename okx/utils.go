package okx

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"strings"

	"github.com/google/uuid"
)

func GzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func FlateDecompress(data []byte) ([]byte, error) {
	return io.ReadAll(flate.NewReader(bytes.NewReader(data)))
}

func GenerateOrderClientId(size int) string {
	uuidStr := strings.Replace(uuid.New().String(), "-", "", 32)
	return uuidStr[0 : size-5]
}
