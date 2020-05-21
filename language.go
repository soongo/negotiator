// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"sort"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

var simpleLanguageRegExp = regexp2.MustCompile("^\\s*([^\\s\\-;]+)(?:-([^\\s;]+))?\\s*(?:;(.*))?$", regexp2.None)

type acceptLanguage struct {
	prefix string
	suffix string
	full   string
	q      float64
	i      int
}

type acceptLanguages []acceptLanguage

func (acs acceptLanguages) filter(f func(ac acceptLanguage) bool) acceptLanguages {
	result := make(acceptLanguages, 0, len(acs))
	for _, ac := range acs {
		if f(ac) {
			result = append(result, ac)
		}
	}
	return result
}

func (acs acceptLanguages) toLanguages() []string {
	result := make([]string, len(acs), len(acs))
	for i, ac := range acs {
		result[i] = ac.full
	}
	return result
}

type acceptLanguageBy func(ac1, ac2 *acceptLanguage) bool

func (by acceptLanguageBy) sort(acs acceptLanguages) {
	as := &acceptLanguageSorter{acs, by}
	sort.Sort(as)
}

type acceptLanguageSorter struct {
	acs acceptLanguages
	by  func(ac1, ac2 *acceptLanguage) bool
}

func (s *acceptLanguageSorter) Len() int {
	return len(s.acs)
}

func (s *acceptLanguageSorter) Swap(i, j int) {
	s.acs[i], s.acs[j] = s.acs[j], s.acs[i]
}

func (s *acceptLanguageSorter) Less(i, j int) bool {
	return s.by(&s.acs[i], &s.acs[j])
}

// PreferredLanguages gets the preferred languages from an Accept-Language header.
// RFC 2616 sec 14.2: no header = *, so you should pass * if no Accept-Language field in header.
func PreferredLanguages(accept string, provided ...string) []string {
	acs := parseAcceptLanguage(accept)

	if len(provided) == 0 {
		// sorted list of all languages
		filteredAcs := acs.filter(isAcceptLanguageQuality)
		acceptLanguageBy(func(ac1, ac2 *acceptLanguage) bool {
			if ac1.q != ac2.q {
				return ac1.q > ac2.q
			}
			return ac1.i < ac2.i
		}).sort(filteredAcs)
		return filteredAcs.toLanguages()
	}

	// sorted list of accepted languages
	priorities := getLanguageSpecificities(provided, acs)
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

// Parses the Accept-Language header to slice with type acceptLanguage.
func parseAcceptLanguage(accept string) acceptLanguages {
	accepts := strings.Split(accept, ",")
	length := len(accepts)
	results := make(acceptLanguages, 0, length)

	for i := 0; i < length; i++ {
		language := parseLanguage(strings.Trim(accepts[i], " "), i)
		if language != nil {
			results = append(results, *language)
		}
	}

	return results
}

// Parse a language from the Accept-Language header.
func parseLanguage(s string, i int) *acceptLanguage {
	match, err := simpleLanguageRegExp.FindStringMatch(s)
	if match == nil || match.GroupCount() == 0 || err != nil {
		return nil
	}

	prefix, suffix, q := match.Groups()[1].String(), match.Groups()[2].String(), 1.0
	full := prefix
	if suffix != "" {
		full += "-" + suffix
	}
	if match.Groups()[3].String() != "" {
		params := strings.Split(match.Groups()[3].String(), ";")
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

	return &acceptLanguage{prefix, suffix, full, q, i}
}

// Get the priority of a language.
func getLanguagePriority(language string, acs acceptLanguages, index int) specificity {
	priority := specificity{o: -1, q: 0, s: 0}

	for i := 0; i < len(acs); i++ {
		spec := languageSpecify(language, acs[i], index)
		if spec != nil {
			s, q, o := priority.s-spec.s, priority.q-spec.q, priority.o-spec.o
			if s < 0 || q < 0 || o < 0 {
				priority = *spec
			}
		}
	}

	return priority
}

// Get the specificity of the language.
func languageSpecify(language string, ac acceptLanguage, index int) *specificity {
	p := parseLanguage(language, index)
	if p == nil {
		return nil
	}

	s := 0
	if strings.ToLower(ac.full) == strings.ToLower(p.full) {
		s |= 4
	} else if strings.ToLower(ac.prefix) == strings.ToLower(p.full) {
		s |= 2
	} else if strings.ToLower(ac.full) == strings.ToLower(p.prefix) {
		s |= 1
	} else if ac.full != "*" {
		return nil
	}
	return &specificity{index, ac.i, ac.q, s}
}

func isAcceptLanguageQuality(ac acceptLanguage) bool {
	return ac.q > 0
}

func getLanguageSpecificities(types []string, acs acceptLanguages) specificities {
	result := make(specificities, len(types), len(types))
	for i, v := range types {
		result[i] = getLanguagePriority(v, acs, i)
	}
	return result
}
