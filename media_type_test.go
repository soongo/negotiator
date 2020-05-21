// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"reflect"
	"testing"
)

var preferredMediaTypeTestObjs = []testObj{
	{
		"text/html",
		nil,
		[]string{"text/html"},
	},
	{
		"text/html, text/*",
		nil,
		[]string{"text/html", "text/*"},
	},
	{
		"text/html, text/plain;q=0.8",
		nil,
		[]string{"text/html", "text/plain"},
	},
	{
		"text/html, application/*;q=0.2, image/jpeg;q=0.8",
		nil,
		[]string{"text/html", "image/jpeg", "application/*"},
	},
	{
		"text/html",
		[]string{"text/*"},
		[]string{},
	},
	{
		"text/*, image/*",
		[]string{"text/html", "image/*"},
		[]string{"image/*", "text/html"},
	},
	{
		"text/*, image/*",
		[]string{"text/*"},
		[]string{"text/*"},
	},
	{
		"text/html, image/jpeg;q=0.8",
		[]string{"*/*"},
		[]string{},
	},
	{
		"text/html;q=0.6, image/jpeg;q=0.8",
		[]string{"*/*"},
		[]string{},
	},
	{
		"text/*;q=0.1, image/*;q=0.1, application/*;q=0.2",
		[]string{"text/*", "image/*", "application/*"},
		[]string{"application/*", "text/*", "image/*"},
	},
	{
		"text/*;q=0.1, image/*;q=0.1, application/*;q=0.2",
		[]string{"text/*", "image/*", "application/json"},
		[]string{"application/json", "text/*", "image/*"},
	},
	{
		"text/*, image/*;q=0.8, application/*;q=0.2",
		[]string{"text/plain", "application/*"},
		[]string{"text/plain", "application/*"},
	},
	{
		"text/*, image/*;q=0.8, application/*;q=0.2",
		[]string{"text/plain", "application/json"},
		[]string{"text/plain", "application/json"},
	},
	{
		"",
		[]string{"text/*", "image/*"},
		[]string{},
	},
	{
		"text/*, image/*;q=0.8, application/json;q=0.2",
		[]string{},
		[]string{"text/*", "image/*", "application/json"},
	},
	{
		"text/*, image/*;q=0.1, application/json;q=0.2",
		[]string{},
		[]string{"text/*", "application/json", "image/*"},
	},
	{
		"*/*",
		[]string{},
		[]string{"*/*"},
	},
	{
		"*/*",
		[]string{"text/html"},
		[]string{"text/html"},
	},
	{
		"*/*, text/*",
		[]string{},
		[]string{"*/*", "text/*"},
	},
	{
		"*/*;q=0.5, text/*",
		[]string{},
		[]string{"text/*", "*/*"},
	},
	{
		"*/*, text/*;q=x",
		[]string{},
		[]string{"*/*"},
	},
	{
		"*/*, text/*;q=x",
		[]string{"text/html"},
		[]string{"text/html"},
	},
	{
		"text/*, application/json",
		[]string{"application/json", "text/plain"},
		[]string{"application/json", "text/plain"},
	},
}

func TestPreferredMediaTypes(t *testing.T) {
	for _, tt := range preferredMediaTypeTestObjs {
		if got := PreferredMediaTypes(tt.accept, tt.provided...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseAcceptMediaType(t *testing.T) {
	tests := []struct {
		s        string
		expected acceptMediaTypes
	}{
		{"text/html", acceptMediaTypes{{"text", "html", map[string]string{}, 1, 0}}},
		{
			"text/html, application/*;q=0.2, image/jpeg;q=0.8",
			acceptMediaTypes{
				{"text", "html", map[string]string{}, 1, 0},
				{"application", "*", map[string]string{}, .2, 1},
				{"image", "jpeg", map[string]string{}, .8, 2},
			},
		},
		{
			"\"text/html, application/*;q=0.2, image/jpeg;q=0.8\"",
			acceptMediaTypes{},
		},
	}
	for _, tt := range tests {
		if got := parseAcceptMediaType(tt.s); !acceptMediaTypeEquals(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseMediaType(t *testing.T) {
	tests := []struct {
		s        string
		i        int
		expected *acceptMediaType
	}{
		{"text/html", 0, &acceptMediaType{"text", "html", map[string]string{}, 1, 0}},
		{"text/html;q=0.8", 1, &acceptMediaType{"text", "html", map[string]string{}, .8, 1}},
		{"text/*", 2, &acceptMediaType{"text", "*", map[string]string{}, 1, 2}},
		{"text/*;q=.8", 3, &acceptMediaType{"text", "*", map[string]string{}, .8, 3}},
		{"*/*;q=0.8", 4, &acceptMediaType{"*", "*", map[string]string{}, .8, 4}},
		{"text/*;p=0.8", 5, &acceptMediaType{"text", "*", map[string]string{"p": "0.8"}, 1, 5}},
		{"text/*;p=\"", 6, &acceptMediaType{"text", "*", map[string]string{"p": ""}, 1, 6}},
		{"text/*;p=\"0.8", 7, &acceptMediaType{"text", "*", map[string]string{"p": "\"0.8"}, 1, 7}},
		{"text/*;p=\"0.8\"", 8, &acceptMediaType{"text", "*", map[string]string{"p": "0.8"}, 1, 8}},
		{"text/*;q=\"0.8\"", 9, &acceptMediaType{"text", "*", map[string]string{}, .8, 9}},
		{"text/html ; q=0.8", 10, &acceptMediaType{"text", "html", map[string]string{}, .8, 10}},
		{"text/html;q=x", 11, nil},
	}
	for _, tt := range tests {
		got := parseMediaType(tt.s, tt.i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestGetMediaTypePriority(t *testing.T) {
	acs := acceptMediaTypes{
		{"text", "html", map[string]string{}, 1, 0},
		{"text", "*", map[string]string{}, .8, 1},
	}
	tests := []struct {
		mediaType string
		acs       acceptMediaTypes
		index     int
		expected  specificity
	}{
		{"text/html", acceptMediaTypes{}, 0, specificity{0, -1, 0, 0}},
		{"text/html", acs, 1, specificity{1, 1, 0.8, 4}},
		{"text/*", acs, 2, specificity{2, 1, .8, 6}},
		{"text/plain", acs, 3, specificity{3, 1, .8, 4}},
		{"image/png", acs, 4, specificity{0, -1, 0, 0}},
		{"image/*", acs, 5, specificity{0, -1, 0, 0}},
		{"*/*", acs, 6, specificity{0, -1, 0, 0}},
	}
	for _, tt := range tests {
		got := getMediaTypePriority(tt.mediaType, tt.acs, tt.index)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestMediaTypeSpecify(t *testing.T) {
	tests := []struct {
		mediaType string
		ac        acceptMediaType
		index     int
		expected  *specificity
	}{
		{
			"text/html",
			acceptMediaType{"text", "html", map[string]string{}, 1, 0},
			0,
			&specificity{0, 0, 1, 6},
		},
		{
			"text/html;q=0.8",
			acceptMediaType{"text", "html", map[string]string{}, .8, 1},
			1,
			&specificity{1, 1, .8, 6},
		},
		{
			"text/*",
			acceptMediaType{"text", "*", map[string]string{}, 1, 2},
			2,
			&specificity{2, 2, 1, 6},
		},
		{
			"text/*;q=0.8",
			acceptMediaType{"text", "*", map[string]string{}, .8, 3},
			3,
			&specificity{3, 3, .8, 6},
		},
		{
			"text/html;p=0.8",
			acceptMediaType{"text", "html", map[string]string{}, .8, 4},
			4,
			&specificity{4, 4, .8, 6},
		},
		{
			"text/html;p=\"",
			acceptMediaType{"text", "html", map[string]string{}, .8, 5},
			5,
			&specificity{5, 5, .8, 6},
		},
		{
			"text/html;p=\"0.8\"",
			acceptMediaType{"text", "html", map[string]string{}, .8, 6},
			6,
			&specificity{6, 6, .8, 6},
		},
		{
			"text/html;q=\"0.8\"",
			acceptMediaType{"text", "html", map[string]string{}, .8, 7},
			7,
			&specificity{7, 7, .8, 6},
		},
		{
			"text/html",
			acceptMediaType{"text", "*", map[string]string{}, 1, 8},
			8,
			&specificity{8, 8, 1, 4},
		},
		{
			"text/*",
			acceptMediaType{"text", "html", map[string]string{}, 1, 9},
			9,
			nil,
		},
		{
			"text/*",
			acceptMediaType{"image", "*", map[string]string{}, 1, 10},
			10,
			nil,
		},
		{
			"text/*",
			acceptMediaType{"*", "*", map[string]string{}, 1, 11},
			11,
			&specificity{11, 11, 1, 2},
		},
		{
			"",
			acceptMediaType{"*", "*", map[string]string{}, 1, 12},
			12,
			nil,
		},
		{
			"text/html",
			acceptMediaType{"*", "*", map[string]string{"foo": "bar"}, 1, 13},
			13,
			nil,
		},
		{
			"text/html",
			acceptMediaType{"*", "*", map[string]string{"foo": "*"}, 1, 14},
			14,
			&specificity{14, 14, 1, 1},
		},
	}
	for i, tt := range tests {
		got := mediaTypeSpecify(tt.mediaType, tt.ac, i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestQuoteCount(t *testing.T) {
	tests := []struct {
		s        string
		expected int
	}{
		{"\"", 1},
		{"\"foo\"", 2},
		{"\"foo\": \"bar\"", 4},
	}
	for _, tt := range tests {
		if got := quoteCount(tt.s); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSplitKeyValuePair(t *testing.T) {
	tests := []struct {
		s        string
		expected []string
	}{
		{"foo", []string{"foo", ""}},
		{"foo=bar", []string{"foo", "bar"}},
	}
	for _, tt := range tests {
		if got := splitKeyValuePair(tt.s); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSplitMediaTypes(t *testing.T) {
	tests := []struct {
		s        string
		expected []string
	}{
		{
			"text/html, application/*;q=0.2, image/jpeg;q=0.8",
			[]string{"text/html", " application/*;q=0.2", " image/jpeg;q=0.8"},
		},
		{
			"\"text/html, application/*;q=0.2, image/jpeg;q=0.8\"",
			[]string{`"text/html, application/*;q=0.2, image/jpeg;q=0.8"`},
		},
	}
	for _, tt := range tests {
		if got := splitMediaTypes(tt.s); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSplitParameters(t *testing.T) {
	tests := []struct {
		s        string
		expected []string
	}{
		{
			"application/*;q=0.2",
			[]string{"application/*", "q=0.2"},
		},
		{
			" application/* ; q=0.2 ",
			[]string{"application/*", "q=0.2"},
		},
		{
			"\"application/*;q=0.2",
			[]string{"\"application/*;q=0.2"},
		},
	}
	for _, tt := range tests {
		if got := splitParameters(tt.s); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func acceptMediaTypeEquals(a, b acceptMediaTypes) bool {
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
