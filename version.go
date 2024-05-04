package ultrastar

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// Version represents the version of an UltraStar song file.
// A version consists of a major, minor and patch version.
type Version struct {
	Major uint // major component
	Minor uint // minor component
	Patch uint // patch component
}

// ParseVersion parses a version number from s.
// A version number is a triplet of positive integers, separated by periods.
// For example "1.0.0" and "15.3.6" are valid versions.
// This function does not support parsing shorter version formats such as "1.0" or "v4".
//
// If s does not contain a valid version, an error is returned.
func ParseVersion(s string) (Version, error) {
	var v Version
	comps := strings.Split(s, ".")
	if len(comps) != 3 {
		return v, fmt.Errorf("version has %d components instead of 3: %s", len(comps), s)
	}
	if m, err := strconv.ParseUint(comps[0], 10, 64); err != nil {
		return v, err
	} else {
		v.Major = uint(m)
	}
	if m, err := strconv.ParseUint(comps[1], 10, 64); err != nil {
		return v, err
	} else {
		v.Minor = uint(m)
	}
	if p, err := strconv.ParseUint(comps[2], 10, 64); err != nil {
		return v, err
	} else {
		v.Patch = uint(p)
	}
	return v, nil
}

// MustParseVersion works like [ParseVersion] but panics if s describes an invalid version.
func MustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// String returns a string representation of v.
// The returned string is a triplet of positive integers separated by dots.
// The resulting string is compatible with [ParseVersion].
func (v Version) String() string {
	return strconv.FormatUint(uint64(v.Major), 10) + "." + strconv.FormatUint(uint64(v.Minor), 10) + "." + strconv.FormatUint(uint64(v.Patch), 10)
}

// MarshalText encodes v into a textual representation.
func (v Version) MarshalText() (text []byte, err error) {
	s := v.String()
	return []byte(s), nil
}

// UnmarshalText decodes v from a textual representation.
// Supported formates are the same as [ParseVersion].
func (v *Version) UnmarshalText(text []byte) (err error) {
	*v, err = ParseVersion(string(text))
	return err
}

// MarshalBinary encodes v into a binary representation.
func (v Version) MarshalBinary() (data []byte, err error) {
	buf := binary.AppendUvarint(nil, uint64(v.Major))
	buf = binary.AppendUvarint(buf, uint64(v.Minor))
	buf = binary.AppendUvarint(buf, uint64(v.Patch))
	return buf, nil
}

// UnmarshalBinary decodes v from a binary representation.
func (v *Version) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	*v = Version{}
	if m, err := binary.ReadUvarint(r); err != nil {
		return err
	} else {
		v.Major = uint(m)
	}
	if m, err := binary.ReadUvarint(r); err != nil {
		return err
	} else {
		v.Minor = uint(m)
	}
	if p, err := binary.ReadUvarint(r); err != nil {
		return err
	} else {
		v.Patch = uint(p)
	}
	return nil
}

// LessThan compares v to v2 and returns true if v is a lower version number than v2.
func (v Version) LessThan(v2 Version) bool {
	return v.Compare(v2) < 0
}

// GreaterThan compares v to v2 and returns true if v is a greater version number than v2.
func (v Version) GreaterThan(v2 Version) bool {
	return v.Compare(v2) > 0
}

// Compare compares v to v2 and returns an integer indicating if v is less than, equal or greater than v2.
func (v Version) Compare(v2 Version) int {
	if v.Major != v2.Major {
		return cmp.Compare(v.Major, v2.Major)
	}
	if v.Minor != v2.Minor {
		return cmp.Compare(v.Minor, v2.Minor)
	}
	return cmp.Compare(v.Patch, v2.Patch)
}

// IsZero returns true if v is the zero value.
func (v Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}
