package hw03frequencyanalysis

import (
	"sort"
	"strings"
	"unicode"
)

func Top10(input string) []string {
	split := strings.Fields(input)
	bib := make(map[string]int)
	for _, v := range split {
		bib[v]++
	}
	words := make([]string, 0, len(bib))
	for word := range bib {
		words = append(words, word)
	}
	sort.Slice(words,
		func(i, j int) bool {
			if bib[words[i]] == bib[words[j]] {
				return words[i] < words[j]
			}
			return bib[words[i]] > bib[words[j]]
		})
	if len(words) > 10 {
		return words[:10]
	}
	return words
}

func Top10Asterisk(input string) []string {
	split := strings.Fields(input)
	bib := make(map[string]int)
	for _, word := range split {
		cleaned := strings.TrimFunc(word,
			func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '-'
			})
		if cleaned == "" || cleaned == "-" {
			continue
		}
		cleaned = strings.ToLower(cleaned)
		bib[cleaned]++
	}

	words := make([]string, 0, len(bib))
	for word := range bib {
		words = append(words, word)
	}

	sort.Slice(words, func(i, j int) bool {
		if bib[words[i]] == bib[words[j]] {
			return words[i] < words[j]
		}
		return bib[words[i]] > bib[words[j]]
	})

	if len(words) > 10 {
		return words[:10]
	}
	return words
}
