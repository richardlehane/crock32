package crock32

import "testing"

func TestDecode(t *testing.T) {
	SetUpper()
	cases := []struct {
		in  string
		exp uint64
	}{
		// 'c' = 12, 'o' = 0 -> 12*32 + 0 = 384.
		{"co", 384},
		{"c0", 384},
		// 'f' = 15, 'i'/'l' = 1 -> 15*32 + 1 = 481.
		{"fi", 481},
		{"fl", 481},
		{"f1", 481},
		// 9,1,O(=0) -> 9*1024 + 32 + 0 = 9248.
		{"91O", 9248},
		{"9-1-0", 9248}, // hypens ignored
	}
	for _, c := range cases {
		got, err := Decode(c.in)
		if err != nil {
			t.Errorf("Decode(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.exp {
			t.Errorf("Decode(%q) = %d, want %d", c.in, got, c.exp)
		}
	}
}

func TestInvalidChar(t *testing.T) {
	for _, in := range []string{"u", "U", "!", " ", "c0u"} {
		if _, err := Decode(in); err == nil {
			t.Errorf("Decode(%q) expected error for invalid character, got nil", in)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	values := []uint64{
		0, 1, 2, 31, 32, 33, 48, 49, 384, 481, 9248, 393505,
		1<<32 - 1, 1 << 32, 1<<63 - 1, ^uint64(0),
	}
	for _, n := range values {
		enc := Encode(n)
		back, err := Decode(enc)
		if err != nil {
			t.Errorf("Decode(Encode(%d)=%q) unexpected error: %v", n, enc, err)
			continue
		}
		if back != n {
			t.Errorf("Decode(Encode(%d)=%q) = %d, want %d", n, enc, back, n)
		}
	}
	for _, n := range values {
		enc := EncodeWithCheck(n)
		back, err := DecodeWithCheck(enc)
		if err != nil {
			t.Errorf("Decode(Encode(%d)=%q) unexpected error: %v", n, enc, err)
			continue
		}
		if back != n {
			t.Errorf("Decode(Encode(%d)=%q) = %d, want %d", n, enc, back, n)
		}
	}
}

func TestEncodeCheck(t *testing.T) {
	cases := []struct {
		in  uint64
		exp string
	}{
		{384, "c0e"},
		{481, "f10"},
		{9248, "910="},
		{67802, "226tj"},
	}
	for _, c := range cases {
		got := EncodeWithCheck(c.in)
		if got != c.exp {
			t.Errorf("Encode with check (%d) = %s, want %s", c.in, got, c.exp)
		}
	}
}
