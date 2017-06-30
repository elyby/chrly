package redis

import (
	"bytes"
	"compress/zlib"
	"io"
	"strings"
)

func buildKey(username string) string {
	return "username:" + strings.ToLower(username)
}

func zlibEncode(str []byte) []byte {
	var buff bytes.Buffer
	writer := zlib.NewWriter(&buff)
	writer.Write(str)
	writer.Close()

	return buff.Bytes()
}

func zlibDecode(bts []byte) ([]byte, error) {
	buff := bytes.NewReader(bts)
	reader, readError := zlib.NewReader(buff)
	if readError != nil {
		return nil, readError
	}

	resultBuffer := new(bytes.Buffer)
	io.Copy(resultBuffer, reader)
	reader.Close()

	return resultBuffer.Bytes(), nil
}
