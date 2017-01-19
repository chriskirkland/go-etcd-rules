package rules

import (
	"regexp"
	"strings"
)

type keyMatcher interface {
	getPrefix() string
	match(string) (keyMatch, bool)
	getPattern() string
	getPrefixesWithConstraints(constraints map[string]constraint) []string
}

type regexKeyMatcher struct {
	regex    *regexp.Regexp
	fieldMap map[string]int
	pattern  string
}

type keyMatch interface {
	GetAttribute(name string) *string
	Format(pattern string) string
}

func (rkm *regexKeyMatcher) getPrefix() string {
	end := strings.Index(rkm.pattern, ":")
	if end == -1 {
		end = len(rkm.pattern)
	}
	return rkm.pattern[0:end]
}

func (rkm *regexKeyMatcher) getPrefixesWithConstraints(constraints map[string]constraint) []string {
	out := []string{}
	firstColon := strings.Index(rkm.pattern, ":")
	if firstColon == -1 {
		out = append(out, rkm.getPrefix())
	} else {
		end := strings.Index(rkm.pattern[firstColon:], "/")
		if end == -1 {
			end = len(rkm.pattern)
		} else {
			end = firstColon + end
		}
		attrName := rkm.pattern[firstColon+1 : end]
		constr, ok := constraints[attrName]
		if !ok {
			out = append(out, rkm.getPrefix())
		} else {
			outPtr := &out
			buildPrefixesFromConstraint(rkm.pattern[:firstColon]+constr.prefix, 0, constr, outPtr)
			out = *outPtr
		}
	}
	return out
}

func buildPrefixesFromConstraint(base string, index int, constr constraint, prefixes *[]string) {
	myChars := constr.chars[index]
	if index+1 == len(constr.chars) {
		// Last set
		for _, char := range myChars {
			newPrefixes := append(*prefixes, base+string(char))
			*prefixes = newPrefixes
		}
	} else {
		for _, char := range myChars {
			newBase := base + string(char)
			buildPrefixesFromConstraint(newBase, index+1, constr, prefixes)
		}
	}
}

func (rkm *regexKeyMatcher) getPattern() string {
	return rkm.pattern
}

type regexKeyMatch struct {
	matchStrings []string
	fieldMap     map[string]int
}

func newKeyMatch(path string, kmr *regexKeyMatcher) *regexKeyMatch {
	results := kmr.regex.FindStringSubmatch(path)
	if results == nil {
		return nil
	}
	km := &regexKeyMatch{
		matchStrings: results,
		fieldMap:     kmr.fieldMap,
	}
	return km
}

func (m *regexKeyMatch) GetAttribute(name string) *string {
	index, ok := m.fieldMap[name]
	if !ok {
		return nil
	}
	result := m.matchStrings[index]
	return &result
}

func (m *regexKeyMatch) Format(pattern string) string {
	return formatWithAttributes(pattern, m)
}
func formatWithAttributes(pattern string, m Attributes) string {
	paths := strings.Split(pattern, "/")
	result := ""
	for _, path := range paths {
		if len(path) == 0 {
			continue
		}
		result = result + "/"
		if strings.HasPrefix(path, ":") {
			attr := m.GetAttribute(path[1:])
			if attr == nil {
				s := path
				attr = &s
			}
			result = result + *attr
		} else {
			result = result + path
		}
	}
	return result
}

// Keep the bool return value, because it's tricky to check for null
// references when dealing with interfaces
func (rkm *regexKeyMatcher) match(path string) (keyMatch, bool) {
	m := newKeyMatch(path, rkm)
	if m == nil {
		return nil, false
	}
	return m, true
}

func newRegexKeyMatcher(pattern string) (*regexKeyMatcher, error) {
	fields, regexString := parsePath(pattern)
	regex, err := regexp.Compile(regexString)
	if err != nil {
		return nil, err
	}
	return &regexKeyMatcher{
		regex:    regex,
		fieldMap: fields,
		pattern:  pattern,
	}, nil
}

type mapAttributes struct {
	values map[string]string
}

func (ma *mapAttributes) GetAttribute(key string) *string {
	value, ok := ma.values[key]
	if !ok {
		return nil
	}
	return &value
}

func (ma *mapAttributes) Format(path string) string {
	return formatWithAttributes(path, ma)
}

func parsePath(pattern string) (map[string]int, string) {
	paths := strings.Split(pattern, "/")
	regex := ""
	fields := make(map[string]int)
	fieldIndex := 1
	for _, path := range paths {
		if len(path) == 0 {
			continue
		}
		regex = regex + "/"
		if strings.HasPrefix(path, ":") {
			regex = regex + "([^\\/:]+)"
			fields[path[1:]] = fieldIndex
			fieldIndex++
		} else {
			regex = regex + path
		}
	}
	return fields, regex
}
