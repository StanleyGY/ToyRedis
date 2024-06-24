package parser

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/stanleygy/toy-redis/app/resp"
)

func parseBulkString(r *bytes.Reader) (string, error) {
	expectedLenByte, err := r.ReadByte()
	if err != nil {
		return "", nil
	}
	expectedLen, err := strconv.Atoi(string(expectedLenByte))
	if err != nil {
		return "", nil
	}

	// Consume \r\n
	r.ReadByte()
	r.ReadByte()

	// Consume string parameters
	str := make([]byte, expectedLen)
	actualLen, err := r.Read(str)
	if err != nil {
		return "", err
	}
	if actualLen != int(expectedLen) {
		return "", errors.New("failed to parse BulkString type: length doesn't match")
	}

	r.ReadByte()
	r.ReadByte()
	return string(str), nil
}

func parseInteger(r *bytes.Reader) (int, error) {
	// Format :[<+|->]<value>\r\n
	rawNum := make([]rune, 0)
	for {
		c, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if c == '\r' {
			break
		}
		rawNum = append(rawNum, rune(c))
	}
	r.ReadByte()
	return strconv.Atoi(string(rawNum))
}

func parseArray(r *bytes.Reader) ([]*resp.RespValue, error) {
	// A sample array: "ECHO hey" is serialized to "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	expectedLenByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	expectedLen, err := strconv.Atoi(string(expectedLenByte))
	if err != nil {
		return nil, err
	}

	// Consume \r\n
	r.ReadByte()
	r.ReadByte()

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
		// Type: Bulk String
		val.BulkStr, err = parseBulkString(r)
	case resp.TypeIntegers:
		val.Int, err = parseInteger(r)
	case resp.TypeArrays:
		// Type: Array
		val.Array, err = parseArray(r)
	}
	return &val, err
}

func Parse(buf []byte) (*resp.RespValue, error) {
	// fmt.Println("cmd ", string(buf))
	r := bytes.NewReader(buf)
	return parseType(r)
}
