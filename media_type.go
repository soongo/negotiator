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

var simpleMediaTypeRegExp = regexp2.MustCompile("^\\s*([^\\s\\/;]+)\\/([^;\\s]+)\\s*(?:;(.*))?$", regexp2.None)

type acceptMediaType struct {
	mainType string
	subtype  string
	params   map[string]string
	q        float64
	i        int
}

type acceptMediaTypes []acceptMediaType

func (acs acceptMediaTypes) filter(f func(ac acceptMediaType) bool) acceptMediaTypes {
	result := make(acceptMediaTypes, 0, len(acs))
	for _, ac := range acs {
		if f(ac) {
			result = append(result, ac)
		}
	}
	return result
}

func (acs acceptMediaTypes) toMediaTypes() []string {
	result := make([]string, len(acs), len(acs))
	for i, ac := range acs {
		result[i] = ac.mainType + "/" + ac.subtype
	}
	return result
}

type acceptMediaTypeBy func(ac1, ac2 *acceptMediaType) bool

func (by acceptMediaTypeBy) sort(acs acceptMediaTypes) {
	as := &acceptMediaTypeSorter{acs, by}
	sort.Sort(as)
}

type acceptMediaTypeSorter struct {
	acs acceptMediaTypes
	by  func(ac1, ac2 *acceptMediaType) bool
}

func (s *acceptMediaTypeSorter) Len() int {
	return len(s.acs)
}

func (s *acceptMediaTypeSorter) Swap(i, j int) {
	s.acs[i], s.acs[j] = s.acs[j], s.acs[i]
}

func (s *acceptMediaTypeSorter) Less(i, j int) bool {
	return s.by(&s.acs[i], &s.acs[j])
}

// PreferredMediaTypes gets the preferred media types from an Accept header.
// RFC 2616 sec 14.2: no header = */*, so you should pass */* if no Accept field in header.
func PreferredMediaTypes(accept string, provided ...string) []string {
	acs := parseAcceptMediaType(accept)

	if len(provided) == 0 {
		// sorted list of all media types
		filteredAcs := acs.filter(isAcceptMediaTypeQuality)
		acceptMediaTypeBy(func(ac1, ac2 *acceptMediaType) bool {
			return ac1.q > ac2.q || ac1.i < ac2.i
		}).sort(filteredAcs)
		return filteredAcs.toMediaTypes()
	}

	priorities := getMediaTypeSpecificities(provided, acs)
	filteredPriorities := priorities.filter(isSpecificityQuality)
	specificityBy(func(s1, s2 *specificity) bool {
		return s1.q > s2.q || s1.s < s2.s || s1.o < s2.o || s1.i < s2.i
	}).sort(filteredPriorities)

	results := make([]string, 0, len(filteredPriorities))
	for _, v := range filteredPriorities {
		i := priorities.indexOf(v)
		if i >= 0 {
			results = append(results, provided[i])
		}
	}

	return results
}

// Parses the Accept header to slice with type acceptMediaType.
func parseAcceptMediaType(accept string) acceptMediaTypes {
	accepts := splitMediaTypes(accept)
	length := len(accepts)
	results := make(acceptMediaTypes, 0, length)

	for i := 0; i < length; i++ {
		mediaType := parseMediaType(strings.Trim(accepts[i], " "), i)
		if mediaType != nil {
			results = append(results, *mediaType)
		}
	}

	return results
}

// Parse a media type from the Accept header.
func parseMediaType(s string, i int) *acceptMediaType {
	match, err := simpleMediaTypeRegExp.FindStringMatch(s)
	if match == nil || match.GroupCount() == 0 || err != nil {
		return nil
	}

	params := make(map[string]string)
	mainType, subType, q := match.Groups()[1].String(), match.Groups()[2].String(), 1.0
	if match.Groups()[3].String() != "" {
		kvps := splitParameters(match.Groups()[3].String())
		arr := make([][]string, len(kvps), len(kvps))
		for i, v := range kvps {
			arr[i] = splitKeyValuePair(v)
		}

		for j := 0; j < len(arr); j++ {
			pair := arr[j]
			key, val := strings.ToLower(pair[0]), pair[1]
			if val != "" && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1:int(math.Max(float64(len(val)-1), 1))]
			}
			if key == "q" {
				q1, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return nil
				}
				q = q1
				break
			}
			params[key] = val
		}
	}

	return &acceptMediaType{mainType, subType, params, q, i}
}

// Get the priority of a media type.
func getMediaTypePriority(mediaType string, acs acceptMediaTypes, index int) specificity {
	priority := specificity{o: -1, q: 0, s: 0}

	for i := 0; i < len(acs); i++ {
		spec := mediaTypeSpecify(mediaType, acs[i], index)
		if spec != nil {
			s, q, o := priority.s-spec.s, priority.q-spec.q, priority.o-spec.o
			if s < 0 || q < 0 || o < 0 {
				priority = *spec
			}
		}
	}

	return priority
}

// Get the specificity of the media type.
func mediaTypeSpecify(mediaType string, ac acceptMediaType, index int) *specificity {
	p := parseMediaType(mediaType, index)
	if p == nil {
		return nil
	}

	s := 0
	if strings.ToLower(ac.mainType) == strings.ToLower(p.mainType) {
		s |= 4
	} else if ac.mainType != "*" {
		return nil
	}

	if strings.ToLower(ac.subtype) == strings.ToLower(p.subtype) {
		s |= 2
	} else if ac.subtype != "*" {
		return nil
	}

	keys := getMapKeys(ac.params)
	if len(keys) > 0 {
		if every(keys, func(k string) bool {
			return ac.params[k] == "*" || strings.ToLower(ac.params[k]) == strings.ToLower(p.params[k])
		}) {
			s |= 1
		} else {
			return nil
		}
	}

	return &specificity{index, ac.i, ac.q, s}
}

func isAcceptMediaTypeQuality(ac acceptMediaType) bool {
	return ac.q > 0
}

func getMediaTypeSpecificities(types []string, acs acceptMediaTypes) specificities {
	result := make(specificities, len(types), len(types))
	for i, v := range types {
		result[i] = getMediaTypePriority(v, acs, i)
	}
	return result
}

// Count the number of quotes in a string.
func quoteCount(s string) int {
	return strings.Count(s, "\"")
}

// Split a key value pair.
func splitKeyValuePair(s string) []string {
	key, val, index := "", "", strings.Index(s, "=")

	if index == -1 {
		key = s
	} else {
		key, val = s[0:index], s[index+1:]
	}

	return []string{key, val}
}

// Split an Accept header into media types.
func splitMediaTypes(accept string) []string {
	accepts := strings.Split(accept, ",")
	length := len(accepts)
	i, j := 1, 0

	for ; i < length; i++ {
		if quoteCount(accepts[j])%2 == 0 {
			j++
			accepts[j] = accepts[i]
		} else {
			accepts[j] += "," + accepts[i]
		}
	}

	accepts = accepts[0 : j+1]

	return accepts
}

// Split a string of parameters.
func splitParameters(str string) []string {
	parameters := strings.Split(str, ";")
	length := len(parameters)
	i, j := 1, 0

	for ; i < length; i++ {
		if quoteCount(parameters[j])%2 == 0 {
			j++
			parameters[j] = parameters[i]
		} else {
			parameters[j] += ";" + parameters[i]
		}
	}

	// trim parameters
	parameters = parameters[0 : j+1]
	length = len(parameters)

	for i = 0; i < length; i++ {
		parameters[i] = strings.Trim(parameters[i], " ")
	}

	return parameters
}

func getMapKeys(m map[string]string) []string {
	i, length := 0, len(m)
	keys := make([]string, length, length)
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func every(arr []string, f func(s string) bool) bool {
	for _, v := range arr {
		if !f(v) {
			return false
		}
	}
	return true
}
