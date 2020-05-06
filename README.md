# negotiator

[![Build Status](https://travis-ci.org/soongo/negotiator.svg)](https://travis-ci.org/soongo/negotiator)
[![codecov](https://codecov.io/gh/soongo/negotiator/branch/master/graph/badge.svg)](https://codecov.io/gh/soongo/negotiator)
[![Go Report Card](https://goreportcard.com/badge/github.com/soongo/negotiator)](https://goreportcard.com/report/github.com/soongo/negotiator)
[![GoDoc](https://godoc.org/github.com/soongo/negotiator?status.svg)](https://godoc.org/github.com/soongo/negotiator)
[![License](https://img.shields.io/badge/MIT-green.svg)](https://opensource.org/licenses/MIT)

An HTTP content negotiator for Go

Thanks to [negotiator](https://github.com/jshttp/negotiator) which is the 
original version written in javascript.

## Installation

To install `negotiator`, you need to install Go and set your Go workspace first.

The first need [Go](https://golang.org/) installed (**version 1.11+ is required**),
then you can use the below Go command to install `negotiator`.

```sh
$ go get -u github.com/soongo/negotiator
```

## Usage

```go
import "github.com/soongo/negotiator"

// The negotiator constructor receives a http.Header object
negotiator := negotiator.New(header)
```

### Accept Negotiation

```go
availableMediaTypes := []string{"text/html", "text/plain", "application/json"}

// Let's say Accept header is 'text/html, application/*;q=0.2, image/jpeg;q=0.8'

negotiator.MediaTypes()
// -> ["text/html", "image/jpeg", "application/*"]

negotiator.MediaTypes(availableMediaTypes)
// -> ["text/html", "application/json"]

negotiator.MediaType(availableMediaTypes)
// -> "text/html"
```

#### Methods

##### MediaType()

Returns the most preferred media type from the client.

##### MediaType(availableMediaType)

Returns the most preferred media type from a list of available media types.

##### MediaTypes()

Returns an array of preferred media types ordered by the client preference.

##### MediaTypes(availableMediaTypes)

Returns an array of preferred media types ordered by priority from a list of
available media types.

### Accept-Language Negotiation

```go
availableLanguages := []string{"en", "es", "fr"}

// Let's say Accept-Language header is 'en;q=0.8, es, pt'

negotiator.Languages()
// -> ["es", "pt", "en"]

negotiator.Languages(availableLanguages)
// -> ["es", "en"]

negotiator.Language(availableLanguages)
// -> "es"
```

#### Methods

##### Language()

Returns the most preferred language from the client.

##### Language(availableLanguages)

Returns the most preferred language from a list of available languages.

##### Languages()

Returns an array of preferred languages ordered by the client preference.

##### Languages(availableLanguages)

Returns an array of preferred languages ordered by priority from a list of
available languages.

### Accept-Charset Negotiation

```go
availableCharsets := []string{"utf-8", "iso-8859-1", "iso-8859-5"}

// Let's say Accept-Charset header is 'utf-8, iso-8859-1;q=0.8, utf-7;q=0.2'

negotiator.Charsets()
// -> ["utf-8", "iso-8859-1", "utf-7"]

negotiator.Charsets(availableCharsets...)
// -> ["utf-8", "iso-8859-1"]

negotiator.Charset(availableCharsets...)
// -> "utf-8"
```

#### Methods

##### Charset()

Returns the most preferred charset from the client.

##### Charset(availableCharsets...)

Returns the most preferred charset from a list of available charsets.

##### Charsets()

Returns an array of preferred charsets ordered by the client preference.

##### Charsets(availableCharsets...)

Returns an array of preferred charsets ordered by priority from a list of
available charsets.

### Accept-Encoding Negotiation

```go
availableEncodings := []string{"identity", "gzip"}

// Let's say Accept-Encoding header is 'gzip, compress;q=0.2, identity;q=0.5'

negotiator.Encodings()
// -> ["gzip", "identity", "compress"]

negotiator.Encodings(availableEncodings...)
// -> ["gzip", "identity"]

negotiator.Encoding(availableEncodings...)
// -> "gzip"
```

#### Methods

##### Encoding()

Returns the most preferred encoding from the client.

##### Encoding(availableEncodings...)

Returns the most preferred encoding from a list of available encodings.

##### Encodings()

Returns an array of preferred encodings ordered by the client preference.

##### Encodings(availableEncodings...)

Returns an array of preferred encodings ordered by priority from a list of
available encodings.
