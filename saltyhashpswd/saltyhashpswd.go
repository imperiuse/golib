package saltyhashpswd

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashAndSalt(pwd []byte, cost int) (string, error) {

	if cost != bcrypt.MinCost && cost != bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	// Use GenerateFromPassword to hash & salt pwd
	hash, err := bcrypt.GenerateFromPassword(pwd, cost)
	if err != nil {
		return "", nil
	}
	return string(hash), nil
}

func ComparePasswords(hashedPwd string, plainPwd []byte) (bool, error) {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
