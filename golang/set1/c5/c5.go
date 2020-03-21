package main

import (
	"encoding/hex"
	"fmt"
)

func multiByteXOR(data, key []byte) []byte {
	res := make([]byte, len(data))

	for i := 0; i < len(data); i++ {
		res[i] = data[i] ^ key[i%len(key)]
	}

	return res
}

func main() {
	input := "Burning 'em, if you ain't quick and nimble\nI go crazy when I hear a cymbal"
	key := "ICE"
	expected := "0b3637272a2b2e63622c2e69692a23693a2a3c6324202d623d63343c2a26226324272765272a282b2f20430a652e2c652a3124333a653e2b2027630c692b20283165286326302e27282f"

	res := hex.EncodeToString(multiByteXOR([]byte(input), []byte(key)))
	if res == expected {
		fmt.Println("Working!")
	}
}
