package core

import (
	"bytes"
	"errors"
	"fmt"
)

func Decode(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	var values []interface{} = make([]interface{}, 0)
	var index int = 0

	for index < len(data) {
		value, delta, err := DecodeOne(data[index:])
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		index += delta
	}
	return values, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	switch data[0] {
	case '*':
		return decodeArray(data)
	case '$':
		return decodeString(data)
	case '+':
		return decodeSimpleString(data)
	case '-':
		return decodeError(data)
	case ':':
		return decodeInteger64(data)
	default:
		return nil, 0, errors.New("unknown data type")
	}

}

func decodeArray(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	pos := 1

	count, delta := readLength(data[pos:])
	pos += delta

	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		result, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		res[i] = result
		pos += delta
	}
	return res, pos, nil
}

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if b < '0' || b > '9' { // #genai: stop at \r when length digits end
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

func decodeString(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}
	pos := 1

	len, delta := readLength(data[pos:])

	pos += delta

	return string(data[pos : pos+len]), pos + len + 2, nil
}

func decodeSimpleString(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	pos := 1
	for pos < len(data) && data[pos] != '\r' {
		pos++
	}

	return string(data[1:pos]), pos + 2, nil
}

func decodeError(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	return decodeSimpleString(data)
}

func decodeInteger64(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, _, err := DecodeOne(data)
	if err != nil {
		return nil, err
	}

	ts := value.([]interface{})
	tokens := make([]string, len(ts))

	for i := range ts {
		tokens[i] = ts[i].(string)
	}
	return tokens, nil
}

func Encode(value interface{}, simple bool) []byte {
	switch v := value.(type) {
	case string:
		if simple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, s := range value.([]string) {
			buf.Write(encodeString(s))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v.Error()))
	}
	return []byte{}
}

func encodeString(s string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}
