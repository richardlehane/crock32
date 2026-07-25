// Copyright 2013 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package crock32 implements Douglas Crockford's Base32 encoding.
//
// Crock32 is useful for "expressing numbers in a form that can be conveniently and accurately transmitted between humans and computer systems."
// See http://www.crockford.com/wrmg/base32.html for details.
// Note: crock32 differs from Crockford in its use of lower-case letters when encoding (decode works for both cases). To change, use: crock32.SetDigits("0123456789ABCDEFGHJKMNPQRSTVWXYZ")
//
// Example:
//
//	i, _ := crock32.Decode("a1j3")
//	s := crock32.Encode(i)
//	fmt.Println(s)
package crock32

import "errors"

const cutoff uint64 = (1<<64-1)/32 + 1

func decodeChar(b byte) byte {
	switch {
	case b == '-':
		return 254
	case b == '*':
		return 32
	case b == '~':
		return 33
	case b == '$':
		return 34
	case b == '=':
		return 35
	case b == 'U', b == 'u':
		return 36
	case b == 'O', b == 'o':
		return 0
	case b == 'L', b == 'l', b == 'I', b == 'i':
		return 1
	case '0' <= b && b <= '9':
		return b - '0'
	case 'a' <= b && b <= 'h':
		return b - 'a' + 10
	case 'A' <= b && b <= 'H':
		return b - 'A' + 10
	case 'j' <= b && b <= 'k':
		return b - 'a' + 9
	case 'J' <= b && b <= 'K':
		return b - 'A' + 9
	case 'm' <= b && b <= 'n':
		return b - 'a' + 8
	case 'M' <= b && b <= 'N':
		return b - 'A' + 8
	case 'p' <= b && b <= 't':
		return b - 'a' + 7
	case 'P' <= b && b <= 'T':
		return b - 'A' + 7
	case 'v' <= b && b <= 'z':
		return b - 'a' + 6
	case 'V' <= b && b <= 'Z':
		return b - 'A' + 6
	default:
		return 255
	}
}

// Decode converts a string matching Douglas Crockford's character set (case insensitive) into an unsigned 64-bit integer.
func Decode(s string) (uint64, error) {
	var n uint64
	for i := 0; i < len(s); i++ {
		v := decodeChar(s[i])
		switch v {
		case 254:
			continue
		case 255:
			return 0, errors.New("crock32.Decode: invalid character " + string(v))
		case 32, 33, 34, 35, 36:
			return 0, errors.New("crock32.Decode: check symbol " + string(v) + " in non-trailing position")
		}
		if n >= cutoff {
			return 0, errors.New("crock32.Decode:" + s + " overflows uint64")
		}
		n = n*32 + uint64(v)
	}
	return n, nil
}

// Decode converts a string matching Douglas Crockford's character set (case insensitive) into an unsigned 64-bit integer.
// Expects a check symbol as the last character and returns an error if check fails.
func DecodeWithCheck(s string) (uint64, error) {
	if len(s) < 2 {
		return 0, errors.New("crock32.DecodeWithCheck: string too short")
	}
	check := decodeChar(s[len(s)-1])
	if check > 36 {
		return 0, errors.New("crock32.DecodeWithCheck: invalid check symbol " + string(check))
	}
	val, err := Decode(s[:len(s)-1])
	if err != nil {
		return 0, err
	}
	if val%37 != uint64(check) {
		return 0, errors.New("crock32.DecodeWithCheck: check symbol " + string(check) + " does not match calculated value")
	}
	return val, nil
}

var digits = "0123456789abcdefghjkmnpqrstvwxyz*~$=u"

// SetDigits allows you to change the encoding alphabet (not the decoding alphabet).
func SetDigits(s string) error {
	if len(s) == 37 {
		digits = s
		return nil
	}
	return errors.New("crock32.SetDigits: character set can be anything but it must be 37 characters long (32 + 5 check digits)")
}

// SetUpper changes the encoding alphabet to upper case (Crockford's preference)
func SetUpper() {
	digits = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
}

const maxuint = 13

// Encode converts a uint64 into a Crockford base32 encoded string
func Encode(n uint64) string {
	var a [maxuint]byte
	i := maxuint
	for n >= 32 {
		i--
		a[i] = digits[n%32]
		n /= 32
	}
	i--
	a[i] = digits[n]
	return string(a[i:])
}

// Encode converts a uint64 into a Crockford base32 encoded string with a trailing check symbol
func EncodeWithCheck(n uint64) string {
	ret := Encode(n)
	check := digits[n%37]
	return ret + string(check)
}
