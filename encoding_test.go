// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"reflect"
	"testing"
)

var preferredEncodingTestObjs = []testObj{
	{
		"gzip",
		nil,
		[]string{"gzip", "identity"},
	},
	{
		"gzip, compress",
		nil,
		[]string{"gzip", "compress", "identity"},
	},
	{
		"gzip, compress;q=0.8",
		nil,
		[]string{"gzip", "compress", "identity"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.2",
		nil,
		[]string{"gzip", "compress", "identity"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.9",
		nil,
		[]string{"gzip", "identity", "compress"},
	},
	{
		"gzip",
		[]string{"gzip", "compress"},
		[]string{"gzip"},
	},
	{
		"gzip, compress",
		[]string{"gzip", "compress"},
		[]string{"gzip", "compress"},
	},
	{
		"gzip, compress",
		[]string{"gzip"},
		[]string{"gzip"},
	},
	{
		"gzip, compress;q=0.8",
		[]string{"gzip", "compress"},
		[]string{"gzip", "compress"},
	},
	{
		"gzip, iso-8859-2;q=0.8",
		[]string{"gzip", "compress"},
		[]string{"gzip"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.2",
		[]string{"gzip", "compress"},
		[]string{"gzip", "compress"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.2",
		[]string{"gzip", "compress", "identity"},
		[]string{"gzip", "compress", "identity"},
	},
	{
		"gzip;q=0.1, compress;q=0.1, identity;q=0.2",
		[]string{"gzip", "compress", "identity"},
		[]string{"identity", "gzip", "compress"},
	},
	{
		"gzip;q=0.1, compress;q=0.2, identity;q=0.3",
		[]string{"gzip", "compress", "identity"},
		[]string{"identity", "compress", "gzip"},
	},
	{
		"gzip;q=0.1, compress;q=0.2, identity;q=0.2",
		[]string{"gzip", "compress", "identity"},
		[]string{"compress", "identity", "gzip"},
	},
	{
		"gzip;q=0.1, compress;q=x, identity;q=0.2",
		[]string{"gzip", "compress", "identity"},
		[]string{"identity", "gzip"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.2",
		[]string{"compress2", "identity"},
		[]string{"identity"},
	},
	{
		"",
		[]string{"gzip", "compress", "identity"},
		[]string{"identity"},
	},
	{
		"gzip, compress;q=0.8, identity;q=0.2",
		[]string{},
		[]string{"gzip", "compress", "identity"},
	},
	{
		"*",
		[]string{},
		[]string{"*"},
	},
	{
		"*",
		[]string{"gzip"},
		[]string{"gzip"},
	},
	{
		"*",
		[]string{"gzip", "compress", "identity"},
		[]string{"gzip", "compress", "identity"},
	},
	{
		"*, gzip",
		[]string{},
		[]string{"*", "gzip"},
	},
	{
		"*, gzip;q=x",
		[]string{},
		[]string{"*"},
	},
	{
		"*, gzip;q=x",
		[]string{"gzip"},
		[]string{"gzip"},
	},
}

func TestPreferredEncodings(t *testing.T) {
	for _, tt := range preferredEncodingTestObjs {
		if got := PreferredEncodings(tt.accept, tt.provided...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseAcceptEncoding(t *testing.T) {
	tests := []struct {
		s        string
		expected acceptEncodings
	}{
		{"gzip", acceptEncodings{
			{"gzip", 1, 0},
			{"identity", 1, 1},
		}},
		{
			"gzip, compress;q=0.8, identity;q=0.2",
			acceptEncodings{
				{"gzip", 1, 0},
				{"compress", .8, 1},
				{"identity", .2, 2},
			},
		},
	}
	for _, tt := range tests {
		if got := parseAcceptEncoding(tt.s); !acceptEncodingEquals(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseEncoding(t *testing.T) {
	tests := []struct {
		s        string
		i        int
		expected *acceptEncoding
	}{
		{"gzip", 0, &acceptEncoding{"gzip", 1, 0}},
		{"compress;q=0.2", 1, &acceptEncoding{"compress", .2, 1}},
		{" compress ; q=0.2 ", 2, &acceptEncoding{"compress", .2, 2}},
		{"gzip;q=x", 3, nil},
	}
	for _, tt := range tests {
		got := parseEncoding(tt.s, tt.i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestGetEncodingPriority(t *testing.T) {
	acs := acceptEncodings{
		{"gzip", 1, 0},
		{"compress", .2, 1},
		{"identity", .5, 2},
	}
	tests := []struct {
		charset  string
		acs      acceptEncodings
		index    int
		expected specificity
	}{
		{"gzip", acceptEncodings{}, 0, specificity{0, -1, 0, 0}},
		{"compress", acs, 1, specificity{1, 1, 0.2, 1}},
		{"identity", acs, 2, specificity{2, 2, 0.5, 1}},
	}
	for _, tt := range tests {
		got := getEncodingPriority(tt.charset, tt.acs, tt.index)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestEncodingSpecify(t *testing.T) {
	tests := []struct {
		encoding string
		ac       acceptEncoding
		index    int
		expected *specificity
	}{
		{
			"gzip",
			acceptEncoding{"gzip", 1, 0},
			0,
			&specificity{0, 0, 1, 1},
		},
		{
			"compress",
			acceptEncoding{"compress", .8, 1},
			1,
			&specificity{1, 1, .8, 1},
		},
		{
			"identity",
			acceptEncoding{"identity", .2, 2},
			2,
			&specificity{2, 2, .2, 1},
		},
		{
			"utf-16",
			acceptEncoding{"utf-32", .3, 3},
			3,
			nil,
		},
		{
			"utf-16",
			acceptEncoding{"*", .4, 4},
			4,
			&specificity{4, 4, .4, 0},
		},
		{
			"*",
			acceptEncoding{"gzip", .5, 5},
			5,
			nil,
		},
		{
			"*",
			acceptEncoding{"*", .6, 6},
			6,
			&specificity{6, 6, .6, 1},
		},
	}
	for i, tt := range tests {
		got := encodingSpecify(tt.encoding, tt.ac, i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func acceptEncodingEquals(a, b acceptEncodings) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}
