package tools

import (
	"io"
	"bytes"
	"compress/zlib"
)

func ZlibEncode(str []byte) []byte {
	var buff bytes.Buffer
	writer := zlib.NewWriter(&buff)
	writer.Write(str)
	writer.Close()

	return buff.Bytes()
}

func ZlibDecode(bts []byte) ([]byte, error) {
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
