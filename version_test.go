package ultrastar

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := map[string]struct {
		raw     string
		want    Version
		wantErr bool
	}{
		"regular":     {"1.0.4", Version{1, 0, 4}, false},
		"incomplete":  {"1.0", Version{}, true},
		"xrange":      {"5.3.x", Version{5, 3, 0}, true},
		"text before": {"foo 4.5.6", Version{}, true},
		"text after":  {"3.5.2 bar", Version{3, 5, 0}, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseVersion(tt.raw)
			if tt.wantErr && err == nil {
				t.Errorf("ParseVersion(%q) = %s, expected error", tt.raw, got)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseVersion(%q) returned an unexpected error: %s", tt.raw, err)
				return
			}
			if tt.want != got {
				t.Errorf("ParseVersion(%q) = %s, expected %s", tt.raw, got, tt.want)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := map[string]struct {
		version Version
		want    string
	}{
		"v0":  {Version{0, 2, 2}, "0.2.2"},
		"v1":  {Version{1, 0, 0}, "1.0.0"},
		"v2":  {Version{2, 4, 1}, "2.4.1"},
		"v23": {Version{23, 9, 9}, "23.9.9"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("String(%v) = %q, expected %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestVersion_IsZero(t *testing.T) {
	tests := map[string]struct {
		version Version
		want    bool
	}{
		"zero":  {Version{}, true},
		"patch": {Version{0, 0, 1}, false},
		"minor": {Version{0, 1, 0}, false},
		"major": {Version{1, 0, 0}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.version.IsZero(); got != tt.want {
				t.Errorf("IsZero(%q) = %t, expected %t", tt.version, got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := map[string]struct {
		v1   Version
		v2   Version
		want int
	}{
		"equal":   {Version{0, 1, 1}, Version{0, 1, 1}, 0},
		"less":    {Version{1, 0, 3}, Version{1, 4, 1}, -1},
		"notless": {Version{2, 5, 1}, Version{1, 6, 7}, 1},
		"greater": {Version{5, 2, 1}, Version{2, 1, 0}, 1},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Compare(tt.v2); got != tt.want {
				t.Errorf("%q.Compare(%q) = %d, expected %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
