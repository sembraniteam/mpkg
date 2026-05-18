package generator

// naturalLess reports whether a < b using natural (version-aware) ordering.
// Digit segments are compared numerically (shorter run = smaller number,
// then lexicographic for equal-length runs). All other segments compare
// lexicographically. No leading-zero ambiguity is expected in package paths.
func naturalLess(a, b string) bool {
	for len(a) > 0 && len(b) > 0 {
		aDigit := a[0] >= '0' && a[0] <= '9'
		bDigit := b[0] >= '0' && b[0] <= '9'

		if aDigit != bDigit {
			// one is a digit, other is not — compare the leading byte directly
			return a[0] < b[0]
		}

		if !aDigit {
			// both non-digit: advance to next digit boundary or end
			ai, bi := 0, 0
			for ai < len(a) && (a[ai] < '0' || a[ai] > '9') {
				ai++
			}
			for bi < len(b) && (b[bi] < '0' || b[bi] > '9') {
				bi++
			}
			as, bs := a[:ai], b[:bi]
			if as != bs {
				return as < bs
			}
			a, b = a[ai:], b[bi:]
		} else {
			// both digit: advance to end of digit run
			ai, bi := 0, 0
			for ai < len(a) && a[ai] >= '0' && a[ai] <= '9' {
				ai++
			}
			for bi < len(b) && b[bi] >= '0' && b[bi] <= '9' {
				bi++
			}
			as, bs := a[:ai], b[:bi]
			// shorter digit string = smaller number (no leading zeros expected)
			if len(as) != len(bs) {
				return len(as) < len(bs)
			}
			if as != bs {
				return as < bs
			}
			a, b = a[ai:], b[bi:]
		}
	}
	return len(a) < len(b)
}
