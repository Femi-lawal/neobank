package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), 10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Hash for 'password':")
	fmt.Println(string(hash))
}
