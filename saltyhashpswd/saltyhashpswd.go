package saltyhashpswd

import (
	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd []byte, cost int) (string, error) {
	if cost < bcrypt.MinCost && cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	// Use GenerateFromPassword to hash & salt pwd
	hash, err := bcrypt.GenerateFromPassword(pwd, cost)
	if err != nil {
		return "", nil
	}

	return string(hash), nil
}

func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	if bcrypt.CompareHashAndPassword([]byte(hashedPwd), plainPwd) != nil {
		return false
	}
	return true
}
