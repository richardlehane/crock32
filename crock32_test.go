package crock32

import "testing"

// TestDecodeNormalization verifies the Crockford Base32 normalization rules:
// O/o decode to the digit 0 and I/i/L/l decode to the digit 1 (numeric values,
// not the ASCII codes 48/49). See https://www.crockford.com/base32.html.
func TestDecodeNormalization(t *testing.T) {
	cases := []struct {
		in  string
		exp uint64
	}{
		// Canonical digits.
		{"0", 0},
		{"1", 1},
		// O/o must equal 0.
		{"O", 0},
		{"o", 0},
		// I/i/L/l must equal 1.
		{"I", 1},
		{"i", 1},
		{"L", 1},
		{"l", 1},
		// Multi-digit cases where a normalized symbol participates.
		// 'c' = 12, 'o' = 0 -> 12*32 + 0 = 384.
		{"co", 384},
		{"c0", 384},
		// 'f' = 15, 'i'/'l' = 1 -> 15*32 + 1 = 481.
		{"fi", 481},
		{"fl", 481},
		{"f1", 481},
		// 9,1,O(=0) -> 9*1024 + 32 + 0 = 9248.
		{"91O", 9248},
		{"910", 9248},
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

// TestDecodeConsistency guards the internal self-contradiction that exposed the
// bug: equivalent symbols must decode to identical values. Before the fix,
// Decode("0")=0 but Decode("O")=48, and Decode("1")=1 but Decode("I")=49.
func TestDecodeConsistency(t *testing.T) {
	pairs := []struct {
		a, b string
	}{
		{"0", "O"},
		{"0", "o"},
		{"1", "I"},
		{"1", "L"},
		{"1", "i"},
		{"1", "l"},
		{"c0", "co"},
		{"f1", "fi"},
		{"f1", "fl"},
		{"910", "91O"},
	}
	for _, p := range pairs {
		va, ea := Decode(p.a)
		vb, eb := Decode(p.b)
		if ea != nil || eb != nil {
			t.Errorf("Decode(%q)=%v/%v or Decode(%q)=%v/%v returned an error", p.a, va, ea, p.b, vb, eb)
			continue
		}
		if va != vb {
			t.Errorf("Decode(%q)=%d and Decode(%q)=%d differ; spec requires them equal", p.a, va, p.b, vb)
		}
	}
}

// TestDecodeHyphensIgnored verifies the Crockford spec rule that hyphens are
// ignored during decoding (they may be inserted for readability). Before the
// fix, a hyphen hit the default error branch and the whole string was rejected.
func TestDecodeHyphensIgnored(t *testing.T) {
	cases := []struct {
		plain, hyphenated string
	}{
		{"c091", "c0-91"},
		{"a1j3", "a1-j3"},
		{"91O", "9-1-O"},
		{"co", "c-o"},
	}
	for _, c := range cases {
		vp, ep := Decode(c.plain)
		vh, eh := Decode(c.hyphenated)
		if ep != nil {
			t.Errorf("Decode(%q) unexpected error: %v", c.plain, ep)
			continue
		}
		if eh != nil {
			t.Errorf("Decode(%q) unexpected error: %v (hyphens must be ignored)", c.hyphenated, eh)
			continue
		}
		if vp != vh {
			t.Errorf("Decode(%q)=%d and Decode(%q)=%d differ; hyphens must be ignored", c.plain, vp, c.hyphenated, vh)
		}
	}
	// Spot value: "c0-91" must equal "c091" = 393505.
	if v, err := Decode("c0-91"); err != nil || v != 393505 {
		t.Errorf("Decode(\"c0-91\") = %d, %v; want 393505, nil", v, err)
	}
}

// TestDecodeInvalidChar ensures symbols outside the Crockford set are still
// rejected. U/u are explicitly excluded by the spec.
func TestDecodeInvalidChar(t *testing.T) {
	for _, in := range []string{"u", "U", "!", " ", "c0u"} {
		if _, err := Decode(in); err == nil {
			t.Errorf("Decode(%q) expected error for invalid character, got nil", in)
		}
	}
}

// TestRoundTrip verifies Decode(Encode(n)) == n for a range of values,
// including values whose encoding contains the digits 0 or 1.
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
}

// TestEncodeUnchanged pins Encode's existing behavior so the fix (which only
// touches Decode) does not regress it.
func TestEncodeUnchanged(t *testing.T) {
	cases := []struct {
		n   uint64
		exp string
	}{
		{0, "0"},
		{1, "1"},
		{32, "10"},
		{384, "c0"},
		{48, "1g"},
		{49, "1h"},
		{9248, "910"},
	}
	for _, c := range cases {
		if got := Encode(c.n); got != c.exp {
			t.Errorf("Encode(%d) = %q, want %q", c.n, got, c.exp)
		}
	}
}
