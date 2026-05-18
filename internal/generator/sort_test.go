package generator

import "testing"

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		// pure text
		{"alpha", "beta", true},
		{"beta", "alpha", false},
		{"alpha", "alpha", false},

		// pure numbers
		{"1", "2", true},
		{"9", "10", true},
		{"10", "9", false},
		{"2", "10", true},
		{"10", "2", false},

		// mixed: text + number
		{"v1", "v2", true},
		{"v2", "v1", false},
		{"v9", "v10", true},
		{"v10", "v9", false},
		{"v1", "v10", true},
		{"v10", "v1", false},
		{"v10", "v10", false},

		// versioned module paths
		{"mpkg/v1", "mpkg/v2", true},
		{"mpkg/v9", "mpkg/v10", true},
		{"mpkg/v10", "mpkg/v9", false},
		{"mpkg/v2", "mpkg/v10", true},
		{"mpkg/v19", "mpkg/v20", true},
		{"mpkg/v20", "mpkg/v19", false},

		// prefix differences
		{"a1", "b1", true},
		{"b1", "a1", false},

		// empty strings
		{"", "a", true},
		{"a", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		got := naturalLess(tt.a, tt.b)
		if got != tt.want {
			t.Errorf(
				"naturalLess(%q, %q) = %v, want %v",
				tt.a,
				tt.b,
				got,
				tt.want,
			)
		}
	}
}
