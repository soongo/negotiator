// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"net/http"
	"net/textproto"
	"strings"
)

// HeaderAcceptCharset is `Accept-Charset`
var HeaderAcceptCharset = textproto.CanonicalMIMEHeaderKey("Accept-Charset")

// HeaderAcceptEncoding is `Accept-Encoding`
var HeaderAcceptEncoding = textproto.CanonicalMIMEHeaderKey("Accept-Encoding")

// HeaderAcceptLanguage is `Accept-Language`
var HeaderAcceptLanguage = textproto.CanonicalMIMEHeaderKey("Accept-Language")

// HeaderAccept is `Accept`
var HeaderAccept = textproto.CanonicalMIMEHeaderKey("Accept")

// Negotiator gets the negotiation info from http header
type Negotiator struct {
	Header http.Header
}

// New creates a Negotiator instance from a header object.
func New(header http.Header) *Negotiator {
	return &Negotiator{header}
}

// Charset gets the most preferred charset from a list of available charsets.
func (n *Negotiator) Charset(available ...string) string {
	return getMostPreferred(n.Charsets(available...))
}

// Charsets gets an array of preferred charsets ordered by priority from a list
// of available charsets.
func (n *Negotiator) Charsets(available ...string) []string {
	// RFC 2616 sec 14.2: no header = *
	return PreferredCharsets(getAccept(n.Header, HeaderAcceptCharset, "*"), available...)
}

// Encoding gets the most preferred encoding from a list of available encodings.
func (n *Negotiator) Encoding(available ...string) string {
	return getMostPreferred(n.Encodings(available...))
}

// Encodings gets an array of preferred encodings ordered by priority from
// a list of available encodings.
func (n *Negotiator) Encodings(available ...string) []string {
	// RFC 2616 sec 14.2: no header = *
	return PreferredEncodings(getAccept(n.Header, HeaderAcceptEncoding, "*"), available...)
}

// Language gets the most preferred language from a list of available languages.
func (n *Negotiator) Language(available ...string) string {
	return getMostPreferred(n.Languages(available...))
}

// Languages gets an array of preferred languages ordered by priority from a list
// of available languages.
func (n *Negotiator) Languages(available ...string) []string {
	// RFC 2616 sec 14.2: no header = *
	return PreferredLanguages(getAccept(n.Header, HeaderAcceptLanguage, "*"), available...)
}

// MediaType gets the most preferred media type from a list of available media types.
func (n *Negotiator) MediaType(available ...string) string {
	return getMostPreferred(n.MediaTypes(available...))
}

// MediaTypes gets an array of preferred mediaTypes ordered by priority from a list
// of available media types.
func (n *Negotiator) MediaTypes(available ...string) []string {
	// RFC 2616 sec 14.2: no header = */*
	return PreferredMediaTypes(getAccept(n.Header, HeaderAccept, "*/*"), available...)
}

func getMostPreferred(accepts []string) string {
	if len(accepts) == 0 {
		return ""
	}
	return accepts[0]
}

func getAccept(h http.Header, key, defaultValue string) string {
	accept, values := defaultValue, getHeaderValues(h, key)
	if values != nil {
		accept = strings.Join(values, ",")
	}
	return accept
}

// The patch of http.Header.Values for go version lower than 1.4
func getHeaderValues(h http.Header, key string) []string {
	if h == nil {
		return nil
	}
	return h[textproto.CanonicalMIMEHeaderKey(key)]
}
