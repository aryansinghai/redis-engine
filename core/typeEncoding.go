package core

import "errors"

func getType(te uint8) uint8 {
	// get the type of the object
	// shift right by 4 bits and then shift left by 4 bits to get the type
	// or same as te & 0b11110000
	// that is we want to keep the type bits and clear the encoding bits
	return (te >> 4) << 4
}

func getEncoding(te uint8) uint8 {
	// get the encoding of the object
	// and keep the encoding bits and clear the type bits
	// or same as te & 0b00001111
	// that is we want to keep the encoding bits and clear the type bits
	return te & 0b00001111
}

func assertType(te uint8, t uint8) error {
	// assert the type of the object
	// if the type is not the same as the expected type, return an error
	if getType(te) != t {
		return errors.New("type mismatch")
	}
	return nil
}

func assertEncoding(te uint8, e uint8) error {
	// assert the encoding of the object
	// if the encoding is not the same as the expected encoding, return an error
	if getEncoding(te) != e {
		return errors.New("encoding mismatch")
	}
	return nil
}
