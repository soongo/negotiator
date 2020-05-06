// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

var (
	dotRegexp       = regexp.MustCompile("\\s*,\\s*")
	testErrorFormat = "got `%v`, expect `%v`"
)

type negotiatorTestObj struct {
	negotiator *Negotiator
	available  []string
	expected   []string
}

func TestNegotiator_Charset(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredCharsetTestObjs, HeaderAcceptCharset) {
		expected := ""
		if len(tt.expected) > 0 {
			expected = tt.expected[0]
		}
		if got := tt.negotiator.Charset(tt.available...); !reflect.DeepEqual(got, expected) {
			t.Errorf(testErrorFormat, got, expected)
		}
	}
}

func TestNegotiator_Charsets(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredCharsetTestObjs, HeaderAcceptCharset) {
		if got := tt.negotiator.Charsets(tt.available...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestNegotiator_Encoding(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredEncodingTestObjs, HeaderAcceptEncoding) {
		expected := ""
		if len(tt.expected) > 0 {
			expected = tt.expected[0]
		}
		if got := tt.negotiator.Encoding(tt.available...); !reflect.DeepEqual(got, expected) {
			t.Errorf(testErrorFormat, got, expected)
		}
	}
}

func TestNegotiator_Encodings(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredEncodingTestObjs, HeaderAcceptEncoding) {
		if got := tt.negotiator.Encodings(tt.available...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestNegotiator_Language(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredLanguageTestObjs, HeaderAcceptLanguage) {
		expected := ""
		if len(tt.expected) > 0 {
			expected = tt.expected[0]
		}
		if got := tt.negotiator.Language(tt.available...); !reflect.DeepEqual(got, expected) {
			t.Errorf(testErrorFormat, got, expected)
		}
	}
}

func TestNegotiator_Languages(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredLanguageTestObjs, HeaderAcceptLanguage) {
		if got := tt.negotiator.Languages(tt.available...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestNegotiator_MediaType(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredMediaTypeTestObjs, HeaderAccept) {
		expected := ""
		if len(tt.expected) > 0 {
			expected = tt.expected[0]
		}
		if got := tt.negotiator.MediaType(tt.available...); !reflect.DeepEqual(got, expected) {
			t.Errorf(testErrorFormat, got, expected)
		}
	}
}

func TestNegotiator_MediaTypes(t *testing.T) {
	for _, tt := range newNegotiatorTestObjs(preferredMediaTypeTestObjs, HeaderAccept) {
		if got := tt.negotiator.MediaTypes(tt.available...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestGetHeaderValues(t *testing.T) {
	charsets := []string{"utf-8", "iso-8859-1;q=0.8"}
	header := http.Header{HeaderAcceptCharset: charsets}
	tests := []struct {
		h        http.Header
		k        string
		expected []string
	}{
		{nil, "Accept-Charset", nil},
		{header, "Accept-Charset", charsets},
		{header, "accept-Charset", charsets},
		{header, "Accept-charset", charsets},
		{header, "accept-charset", charsets},
		{header, "ACCEPT-CHARSET", charsets},
		{header, "ACCEPT-CHARSET", charsets},
	}
	for _, tt := range tests {
		if got := getHeaderValues(tt.h, tt.k); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func newNegotiatorTestObjs(arr []testObj, k string) []negotiatorTestObj {
	results := make([]negotiatorTestObj, len(arr)+1, len(arr)+1)
	for i, obj := range arr {
		header := http.Header{k: dotRegexp.Split(obj.accept, -1)}
		results[i] = negotiatorTestObj{New(header), obj.provided, obj.expected}
		if i == len(arr)-1 {
			header = http.Header{}
			provided := []string{"x-1/x", "x-2/x"}
			results[i+1] = negotiatorTestObj{New(header), provided, provided}
		}
	}
	return results
}
