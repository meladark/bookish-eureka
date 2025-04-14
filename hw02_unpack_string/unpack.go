package hw02unpackstring

import (
	"errors"
	"unicode"
)

var ErrInvalidString = errors.New("некорректная строка")

func Unpack(input string) (string, error) {
	runes := []rune(input)
	var res []rune
	// Предыдущий символ в строке
	var buff rune
	// Символ для печати
	var symbol rune
	length := len(runes)
	for i := range length {
		char := runes[i]
		if i == 0 {
			if !unicode.IsDigit(char) {
				buff = char
				symbol = char
				continue
			}
			return "", ErrInvalidString
		}
		if buff == '\\' {
			if unicode.IsDigit(char) || char == '\\' {
				symbol = char
			} else {
				return "", ErrInvalidString
			}
			buff = '!'
			continue
		}
		if unicode.IsDigit(char) {
			if !unicode.IsDigit(buff) {
				// Хинт из с
				num := int(char - '0')
				for range num {
					res = append(res, symbol)
				}
				buff = char
			} else {
				return "", ErrInvalidString
			}
		} else {
			if !unicode.IsDigit(buff) {
				res = append(res, symbol)
				symbol = char
				buff = char
				continue
			}
			symbol = char
			buff = char
		}
	}

	if !unicode.IsDigit(buff) && length != 0 {
		res = append(res, symbol)
	}

	return string(res), nil
}
