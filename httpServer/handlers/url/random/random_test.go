package random

import (
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	t.Run("Length of generated string", func(t *testing.T) {
		length := 10
		result := RandomString(length)
		if len(result) != length {
			t.Errorf("wanted %d, got %d", length, len(result))
		}
	})
	t.Run("Charset validity", func(t *testing.T) {
		length := 20
		result := RandomString(length)
		for _, char := range result {
			if !strings.ContainsRune(charset, char) {
				t.Errorf("generated char %d is not in given charset %s", char, charset)
			}
		}
	})
	t.Run("Random uniqness test", func(t *testing.T) {
		seen := make(map[string]bool)
		length := 10
		numStrings := 1000

		for i := 0; i < numStrings; i++ {
			result := RandomString(length)
			if seen[result] == true {
				t.Errorf("duplicate string generated: %s", result)
			}
			seen[result] = true
		}
	})

}
