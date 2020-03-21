package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
)

func main () {
	data, err := hex.DecodeString("49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d")
	if err != nil {
		log.Fatal(err)
	}

	str := base64.StdEncoding.EncodeToString(data)
	fmt.Println(str)
}