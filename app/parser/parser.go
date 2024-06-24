package parser

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/stanleygy/toy-redis/app/resp"
)

var (
	ErrInvalidArgs = errors.New("invalid args")
)

func readUntilLineBreak(r *bytes.Reader) ([]byte, error) {
	buf := make([]byte, 0)

	for {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c == '\r' {
			r.ReadByte()
			break
		}
		buf = append(buf, c)
	}
	return buf, nil
}

func parseBulkString(r *bytes.Reader) (string, error) {
	// Get length of string
	rawExpectedNumBytes, err := readUntilLineBreak(r)
	if err != nil {
		return "", err
	}
	expectedNumBytes, err := strconv.Atoi(string(rawExpectedNumBytes))
	if err != nil {
		return "", err
	}
	// Get actual string
	rawStr, err := readUntilLineBreak(r)
	if err != nil {
		return "", err
	}
	if len(rawStr) != expectedNumBytes {
		return "", ErrInvalidArgs
	}
	return string(rawStr), nil
}

func parseInteger(r *bytes.Reader) (int, error) {
	// Format :[<+|->]<value>\r\n
	rawNum, err := readUntilLineBreak(r)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(rawNum))
}

func parseArray(r *bytes.Reader) ([]*resp.RespValue, error) {
	// A sample array: "ECHO hey" is serialized to "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	expectedLenByte, err := readUntilLineBreak(r)
	if err != nil {
		return nil, err
	}
	expectedLen, err := strconv.Atoi(string(expectedLenByte))
	if err != nil {
		return nil, err
	}

	vals := make([]*resp.RespValue, expectedLen)
	for i := 0; i < expectedLen; i++ {
		vals[i], err = parseType(r)
		if err != nil {
			return nil, err
		}
	}
	return vals, nil
}

func parseType(r *bytes.Reader) (*resp.RespValue, error) {
	t, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	var val resp.RespValue
	val.DataType = string(t)

	switch val.DataType {
	case resp.TypeBulkStrings:
		val.BulkStr, err = parseBulkString(r)
	case resp.TypeIntegers:
		val.Int, err = parseInteger(r)
	case resp.TypeArrays:
		val.Array, err = parseArray(r)
	}
	return &val, err
}

func Parse(buf []byte) (*resp.RespValue, error) {
	r := bytes.NewReader(buf)
	return parseType(r)
}
