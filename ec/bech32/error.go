// Copyright (c) 2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32

import (
	"fmt"
)

// ErrMixedCase is returned when the bech32 string has both lower and uppercase
// characters.
type ErrMixedCase struct{}

func (err ErrMixedCase) Error() string {
	return "string not all lowercase or all uppercase"
}

// ErrInvalidBitGroups is returned when conversion is attempted between byte
// slices using bit-per-element of unsupported value.
type ErrInvalidBitGroups struct{}

func (err ErrInvalidBitGroups) Error() string {
	return "only bit groups between 1 and 8 allowed"
}

// ErrInvalidIncompleteGroup is returned when then byte slice used as input has
// data of wrong length.
type ErrInvalidIncompleteGroup struct{}

func (err ErrInvalidIncompleteGroup) Error() string {
	return "invalid incomplete group"
}

// ErrInvalidLength is returned when the bech32 string has an invalid length
// given the BIP-173 defined restrictions.
type ErrInvalidLength int

func (err ErrInvalidLength) Error() string {
	return fmt.Sprintf("invalid bech32 string length %d", int(err))
}

// ErrInvalidCharacter is returned when the bech32 string has a character
// outside the range of the supported charset.
type ErrInvalidCharacter rune

func (err ErrInvalidCharacter) Error() string {
	return fmt.Sprintf("invalid character in string: '%c'", rune(err))
}

// ErrInvalidSeparatorIndex is returned when the separator character '1' is
// in an invalid position in the bech32 string.
type ErrInvalidSeparatorIndex int

func (err ErrInvalidSeparatorIndex) Error() string {
	return fmt.Sprintf("invalid separator index %d", int(err))
}

// ErrNonCharsetChar is returned when a character outside of the specific
// bech32 charset is used in the string.
type ErrNonCharsetChar rune

func (err ErrNonCharsetChar) Error() string {
	return fmt.Sprintf("invalid character not part of charset: %v", int(err))
}

// ErrInvalidChecksum is returned when the extracted checksum of the string
// is different than what was expected. Both the original version, as well as
// the new bech32m checksum may be specified.
type ErrInvalidChecksum struct {
	Expected  string
	ExpectedM string
	Actual    string
}

func (err ErrInvalidChecksum) Error() string {
	return fmt.Sprintf("invalid checksum (expected (bech32=%v, "+
		"bech32m=%v), got %v)", err.Expected, err.ExpectedM, err.Actual)
}

// ErrInvalidDataByte is returned when a byte outside the range required for
// conversion into a string was found.
type ErrInvalidDataByte byte

func (err ErrInvalidDataByte) Error() string {
	return fmt.Sprintf("invalid data byte: %v", byte(err))
}
