package saltyhashpswd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestHashAndSalt(t *testing.T) {
	plainPasswords := []string{"1", "12345678", strings.Repeat("x", 100)}
	costs := []int{0, bcrypt.MinCost, bcrypt.DefaultCost} // bcrypt.MaxCost}

	for i, plainPassword := range plainPasswords {
		for j, cost := range costs {
			_, err := HashAndSalt([]byte(plainPassword), cost)
			assert.Nil(t, err, "HashAndSalt problem in test case", i, j)
		}
	}
}

func TestComparePasswords(t *testing.T) {
	plainPasswords := []string{"", "12345678", strings.Repeat("x", 100500)}
	costs := []int{0, bcrypt.MinCost, bcrypt.DefaultCost} //, bcrypt.MaxCost}
	for _, plainPassword := range plainPasswords {
		for _, cost := range costs {
			r, _ := HashAndSalt([]byte(plainPassword), cost)
			assert.True(t, ComparePasswords(r, []byte(plainPassword)))
		}
	}
}
