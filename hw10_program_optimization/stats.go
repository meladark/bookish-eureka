package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type User struct {
	Email string `json:"email"`
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		var user User
		line := scanner.Bytes()
		if err := json.Unmarshal(line, &user); err != nil {
			continue
		}
		email := user.Email
		at := strings.LastIndexByte(email, '@')
		if at == -1 || at == len(email)-1 {
			continue
		}
		domainPart := email[at+1:]
		if strings.HasSuffix(domainPart, domain) {
			domainLower := strings.ToLower(domainPart)
			result[domainLower]++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
