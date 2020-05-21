// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

var simpleEncodingRegExp = regexp2.MustCompile("^\\s*([^\\s;]+)\\s*(?:;(.*))?$", regexp2.None)

type acceptEncoding struct {
	encoding string
	q        float64
	i        int
}

type acceptEncodings []acceptEncoding

func (acs acceptEncodings) filter(f func(ac acceptEncoding) bool) acceptEncodings {
	result := make(acceptEncodings, 0, len(acs))
	for _, ac := range acs {
		if f(ac) {
			result = append(result, ac)
		}
	}
	return result
}

func (acs acceptEncodings) toEncodings() []string {
	result := make([]string, len(acs), len(acs))
	for i, ac := range acs {
		result[i] = ac.encoding
	}
	return result
}

type acceptEncodingBy func(ac1, ac2 *acceptEncoding) bool

func (by acceptEncodingBy) sort(acs acceptEncodings) {
	as := &acceptEncodingSorter{acs, by}
	sort.Sort(as)
}

type acceptEncodingSorter struct {
	acs acceptEncodings
	by  func(ac1, ac2 *acceptEncoding) bool
}

func (s *acceptEncodingSorter) Len() int {
	return len(s.acs)
}

func (s *acceptEncodingSorter) Swap(i, j int) {
	s.acs[i], s.acs[j] = s.acs[j], s.acs[i]
}

func (s *acceptEncodingSorter) Less(i, j int) bool {
	return s.by(&s.acs[i], &s.acs[j])
}

// PreferredEncodings gets the preferred encodings from an Accept-Encoding header.
func PreferredEncodings(accept string, provided ...string) []string {
	acs := parseAcceptEncoding(accept)

	if len(provided) == 0 {
		// sorted list of all encodings
		filteredAcs := acs.filter(isAcceptEncodingQuality)
		acceptEncodingBy(func(ac1, ac2 *acceptEncoding) bool {
			if ac1.q != ac2.q {
				return ac1.q > ac2.q
			}
			return ac1.i < ac2.i
		}).sort(filteredAcs)
		return filteredAcs.toEncodings()
	}

	// sorted list of accepted charsets
	priorities := getEncodingSpecificities(provided, acs)
	filteredPriorities := priorities.filter(isSpecificityQuality)
	specificityBy(compareSpecs).sort(filteredPriorities)

	results := make([]string, 0, len(filteredPriorities))
	for _, v := range filteredPriorities {
		i := priorities.indexOf(v)
		if i >= 0 {
			results = append(results, provided[i])
		}
	}

	return results
}

// Parses the Accept-Encoding header to slice with type acceptEncoding.
func parseAcceptEncoding(accept string) acceptEncodings {
	accepts, hasIdentity, minQuality := strings.Split(accept, ","), false, 1.0
	length := len(accepts)
	results := make(acceptEncodings, 0, length+1)

	for i := 0; i < length; i++ {
		encoding := parseEncoding(strings.Trim(accepts[i], " "), i)
		if encoding != nil {
			results = append(results, *encoding)
			spec := encodingSpecify("identity", *encoding, 0)
			hasIdentity = hasIdentity || spec != nil
			minQuality = math.Min(minQuality, encoding.q)
		}
	}

	if !hasIdentity {
		results = append(results, acceptEncoding{"identity", minQuality, length})
	}

	return results
}

// Parse an encoding from the Accept-Encoding header.
func parseEncoding(s string, i int) *acceptEncoding {
	match, err := simpleEncodingRegExp.FindStringMatch(s)
	if match == nil || match.GroupCount() == 0 || err != nil {
		return nil
	}

	encoding, q := match.Groups()[1].String(), 1.0
	if match.Groups()[2].String() != "" {
		params := strings.Split(match.Groups()[2].String(), ";")
		for j := 0; j < len(params); j++ {
			p := strings.Split(strings.Trim(params[j], " "), "=")
			if p[0] == "q" {
				q1, err := strconv.ParseFloat(p[1], 64)
				if err != nil {
					return nil
				}
				q = q1
				break
			}
		}
	}

	return &acceptEncoding{encoding, q, i}
}

// Get the priority of an encoding.
func getEncodingPriority(encoding string, acs acceptEncodings, index int) specificity {
	priority := specificity{o: -1, q: 0, s: 0}

	for i := 0; i < len(acs); i++ {
		spec := encodingSpecify(encoding, acs[i], index)
		if spec != nil {
			s, q, o := priority.s-spec.s, priority.q-spec.q, priority.o-spec.o
			if s < 0 || q < 0 || o < 0 {
				priority = *spec
			}
		}
	}

	return priority
}

// Get the specificity of the encoding.
func encodingSpecify(encoding string, ac acceptEncoding, index int) *specificity {
	s := 0
	if strings.ToLower(ac.encoding) == strings.ToLower(encoding) {
		s |= 1
	} else if ac.encoding != "*" {
		return nil
	}
	return &specificity{index, ac.i, ac.q, s}
}

func isAcceptEncodingQuality(ac acceptEncoding) bool {
	return ac.q > 0
}

func getEncodingSpecificities(types []string, acs acceptEncodings) specificities {
	result := make(specificities, len(types), len(types))
	for i, v := range types {
		result[i] = getEncodingPriority(v, acs, i)
	}
	return result
}
