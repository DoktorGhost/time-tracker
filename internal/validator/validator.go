package validator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

func ValidateDigits(input string, ln int) error {
	if len(input) != ln {
		return errors.New("input length must be " + strconv.Itoa(ln) + " characters")
	}
	for _, ch := range input {
		if !unicode.IsDigit(ch) {
			return errors.New("input character is not a digit")
		}
	}
	return nil
}

func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"

	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		sb.WriteByte(letters[idx.Int64()])
	}

	return sb.String()
}

func SecondToString(seconds int) string {
	min := seconds / 60
	hours := min / 60
	minutes := min - (hours * 60)

	return fmt.Sprintf("%02d ч %02d м", hours, minutes)

}
