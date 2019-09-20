package saltyhashpswd

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// go test -covermode=count -coverprofile=coverage.cov && go tool cover -html=coverage.cov

func TestHashAndSalt(t *testing.T) {
	plainPasswords := []string{"1", "12345678", strings.Repeat("x", 100)}
	costs := []int{0, bcrypt.MinCost, bcrypt.DefaultCost} // bcrypt.MaxCost}

	for i, plainPassword := range plainPasswords {
		for j, cost := range costs {
			_, err := HashAndSalt([]byte(plainPassword), cost)
			if err != nil {
				t.Errorf("HashAndSalt problem in test case %d:%d: Error: %v", i, j, err)
			}
		}
	}
}

func TestComparePasswords(t *testing.T) {
	plainPasswords := []string{"", "12345678", strings.Repeat("x", 100500)}
	costs := []int{0, bcrypt.MinCost, bcrypt.DefaultCost} //, bcrypt.MaxCost}
	for i, plainPassword := range plainPasswords {
		for j, cost := range costs {
			r, _ := HashAndSalt([]byte(plainPassword), cost)
			if r, err := ComparePasswords(r, []byte(plainPassword)); err != nil || !r {
				t.Errorf("ComparePasswords wrong result in test case %d:%d: Error: %v", i, j, err)
			}
		}
	}
}
