// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"reflect"
	"testing"
)

var preferredLanguageTestObjs = []testObj{
	{
		"zh",
		nil,
		[]string{"zh"},
	},
	{
		"zh, en",
		nil,
		[]string{"zh", "en"},
	},
	{
		"zh, en;q=0.8",
		nil,
		[]string{"zh", "en"},
	},
	{
		"zh, en;q=0.8, fr;q=0.2",
		nil,
		[]string{"zh", "en", "fr"},
	},
	{
		"zh, en;q=0.8, fr;q=0.9",
		nil,
		[]string{"zh", "fr", "en"},
	},
	{
		"zh",
		[]string{"zh", "en"},
		[]string{"zh"},
	},
	{
		"zh, en",
		[]string{"zh", "en"},
		[]string{"zh", "en"},
	},
	{
		"zh, en",
		[]string{"zh"},
		[]string{"zh"},
	},
	{
		"zh, en;q=0.8",
		[]string{"zh", "en"},
		[]string{"zh", "en"},
	},
	{
		"zh, iso-8859-2;q=0.8",
		[]string{"zh", "en"},
		[]string{"zh"},
	},
	{
		"zh, en;q=0.8, fr;q=0.2",
		[]string{"zh", "en"},
		[]string{"zh", "en"},
	},
	{
		"zh, en;q=0.8, fr;q=0.2",
		[]string{"zh", "en", "fr"},
		[]string{"zh", "en", "fr"},
	},
	{
		"zh;q=0.1, en;q=0.1, fr;q=0.2",
		[]string{"zh", "en", "fr"},
		[]string{"fr", "zh", "en"},
	},
	{
		"zh;q=0.1, en;q=0.2, fr;q=0.3",
		[]string{"zh", "en", "fr"},
		[]string{"fr", "en", "zh"},
	},
	{
		"zh;q=0.1, en;q=0.2, fr;q=0.2",
		[]string{"zh", "en", "fr"},
		[]string{"en", "fr", "zh"},
	},
	{
		"zh;q=0.1, en;q=x, fr;q=0.2",
		[]string{"zh", "en", "fr"},
		[]string{"fr", "zh"},
	},
	{
		"zh, en;q=0.8, fr;q=0.2",
		[]string{"en2", "fr"},
		[]string{"fr"},
	},
	{
		"",
		[]string{"zh", "en", "fr"},
		[]string{},
	},
	{
		"zh, en;q=0.8, fr;q=0.2",
		[]string{},
		[]string{"zh", "en", "fr"},
	},
	{
		"*",
		[]string{},
		[]string{"*"},
	},
	{
		"*",
		[]string{"zh"},
		[]string{"zh"},
	},
	{
		"*",
		[]string{"zh", "en", "fr"},
		[]string{"zh", "en", "fr"},
	},
	{
		"*, zh",
		[]string{},
		[]string{"*", "zh"},
	},
	{
		"*, zh;q=x",
		[]string{},
		[]string{"*"},
	},
	{
		"*, zh;q=x",
		[]string{"zh"},
		[]string{"zh"},
	},
}

func TestPreferredLanguages(t *testing.T) {
	for _, tt := range preferredLanguageTestObjs {
		if got := PreferredLanguages(tt.accept, tt.provided...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		s        string
		expected acceptLanguages
	}{
		{"zh", acceptLanguages{{"zh", "", "zh", 1, 0}}},
		{
			"zh, en;q=0.8, fr;q=0.6",
			acceptLanguages{
				{"zh", "", "zh", 1, 0},
				{"en", "", "en", .8, 1},
				{"fr", "", "fr", .6, 2},
			},
		},
		{
			"zh-CN, en-US;q=0.8, fr;q=0.6",
			acceptLanguages{
				{"zh", "CN", "zh-CN", 1, 0},
				{"en", "US", "en-US", .8, 1},
				{"fr", "", "fr", .6, 2},
			},
		},
	}
	for _, tt := range tests {
		if got := parseAcceptLanguage(tt.s); !acceptLanguageEquals(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		s        string
		i        int
		expected *acceptLanguage
	}{
		{"zh", 0, &acceptLanguage{"zh", "", "zh", 1, 0}},
		{"zh-CN", 1, &acceptLanguage{"zh", "CN", "zh-CN", 1, 1}},
		{"zh-CN;q=0.8", 2, &acceptLanguage{"zh", "CN", "zh-CN", .8, 2}},
		{"en;q=0.8", 3, &acceptLanguage{"en", "", "en", .8, 3}},
		{" en ; q=0.2 ", 4, &acceptLanguage{"en", "", "en", .2, 4}},
		{"en;q=x", 5, nil},
	}
	for _, tt := range tests {
		got := parseLanguage(tt.s, tt.i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestGetLanguagePriority(t *testing.T) {
	acs := acceptLanguages{
		{"zh", "", "zh", 1, 0},
		{"en", "", "en", .8, 1},
	}
	acs2 := acceptLanguages{
		{"zh", "CN", "zh-CN", 1, 0},
		{"en", "US", "en-US", .8, 1},
	}
	tests := []struct {
		language string
		acs      acceptLanguages
		index    int
		expected specificity
	}{
		{"zh", acceptLanguages{}, 0, specificity{0, -1, 0, 0}},
		{"en", acs, 1, specificity{1, 1, 0.8, 4}},
		{"zh-CN", acs, 2, specificity{2, 0, 1, 1}},
		{"en-US", acs, 3, specificity{3, 1, 0.8, 1}},
		{"zh", acs2, 0, specificity{0, 0, 1, 2}},
		{"en", acs2, 1, specificity{1, 1, 0.8, 2}},
		{"zh-CN", acs2, 2, specificity{2, 0, 1, 4}},
		{"en-US", acs2, 3, specificity{3, 1, 0.8, 4}},
	}
	for _, tt := range tests {
		got := getLanguagePriority(tt.language, tt.acs, tt.index)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestLanguageSpecify(t *testing.T) {
	tests := []struct {
		language string
		ac       acceptLanguage
		index    int
		expected *specificity
	}{
		{
			"zh",
			acceptLanguage{"zh", "", "zh", 1, 0},
			0,
			&specificity{0, 0, 1, 4},
		},
		{
			"zh-CN",
			acceptLanguage{"zh", "CN", "zh-CN", .8, 1},
			1,
			&specificity{1, 1, .8, 4},
		},
		{
			"en",
			acceptLanguage{"en", "", "en", .2, 2},
			2,
			&specificity{2, 2, .2, 4},
		},
		{
			"en-US",
			acceptLanguage{"en", "US", "en-US", .3, 3},
			3,
			&specificity{3, 3, .3, 4},
		},
		{
			"fr",
			acceptLanguage{"*", "", "*", .4, 4},
			4,
			&specificity{4, 4, .4, 0},
		},
		{
			"*",
			acceptLanguage{"fr", "", "fr", .5, 5},
			5,
			nil,
		},
		{
			"*",
			acceptLanguage{"*", "", "*", .6, 6},
			6,
			&specificity{6, 6, .6, 4},
		},
		{
			"",
			acceptLanguage{"*", "", "*", .6, 6},
			7,
			nil,
		},
	}
	for i, tt := range tests {
		got := languageSpecify(tt.language, tt.ac, i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func acceptLanguageEquals(a, b acceptLanguages) bool {
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
