package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func main () {
	data, err := hex.DecodeString("1c0111001f010100061a024b53535009181c")
	if err != nil {
		log.Fatal(err)
	}

	key, err := hex.DecodeString("686974207468652062756c6c277320657965")
	if err != nil {
		log.Fatal(err)
	}
	
	var output []byte
	for i, v := range data {
		output = append(output, v ^ key[i])
	}

	fmt.Println(hex.EncodeToString(output))
}