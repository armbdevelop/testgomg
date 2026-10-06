package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	argsCount  = 3
	tokenHours = 24
)

func main() {
	if len(os.Args) != argsCount {
		fmt.Println("usage: gentoken <jwt-secret> <user-id>")
		os.Exit(1)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   os.Args[2],
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenHours * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	signed, err := token.SignedString([]byte(os.Args[1]))
	if err != nil {
		fmt.Println("sign:", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
