package core

type Obj struct {
	TypeEncoding   uint8
	Value          interface{}
	LastAccessedAt uint32
}

// type encoding is a 8 bit integer
// the first 4 bits are the type
// the last 4 bits are the encoding
// the type is the type of the object
// the encoding is the encoding of the object
// the type and encoding are combined to form the type encoding
// the type encoding is used to determine the type and encoding of the object
// the type encoding is used to determine the type and encoding of the object

var OBJ_TYPE_STRING uint8 = 0 << 4

// String Encoding
// string can be encoded in 3 ways:
// 1. raw string
// 2. integer
// 3. embstr string
// raw string is a simple string that is not encoded
// integer is a integer that is encoded as a string
// embstr string is a string that is encoded as a string if the string is small enough
var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8
