package utils

import "strings"

func StripExtraSpacesAndTabs(s string) string {
	s = strings.Replace(s, " ", "", countSpaces(s))
	return strings.Replace(s, "\t", "", 2)
}

func CountSpacesAndTabs(s string) int {
	count := 0
Loop:
	for _, c := range s {
		switch c {
		case ' ':
			count++
		case '\t':
			count++
		default:
			break Loop
		}
	}

	return count
}

func countSpaces(s string) int {
	count := 0
Loop:
	for _, c := range s {
		switch c {
		case ' ':
			count++
		case '\t':
			continue
		default:
			break Loop
		}
	}

	return count
}
