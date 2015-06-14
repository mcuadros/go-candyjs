package candyjs

import "strings"

func isExported(name string) bool {
	return nameToJavaScript(name) != name
}

func nameToJavaScript(name string) string {
	var toLower, keep string
	for _, c := range name {
		if c >= 'A' && c <= 'Z' && len(keep) == 0 {
			toLower += string(c)
		} else {
			keep += string(c)
		}
	}

	lc := len(toLower)
	if lc > 1 && lc != len(name) {
		keep = toLower[lc-1:] + keep
		toLower = toLower[:lc-1]

	}

	return strings.ToLower(toLower) + keep
}

func nameToGo(name string) []string {
	if name[0] >= 'A' && name[0] <= 'Z' {
		return nil
	}

	var toUpper, keep string
	for _, c := range name {
		if c >= 'a' && c <= 'z' && len(keep) == 0 {
			toUpper += string(c)
		} else {
			keep += string(c)
		}
	}

	return []string{
		strings.Title(name),
		strings.ToUpper(toUpper) + keep,
	}
}
