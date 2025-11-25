package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sort"
)

// FrequencyTable approximate character frequency table of the English language
var FrequencyTable = map[byte]float64{
	'a': 0.0651738, 'b': 0.0124248, 'c': 0.0217339, 'd': 0.0349835, 'e': 0.1041442, 'f': 0.0197881, 'g': 0.0158610,
	'h': 0.0492888, 'i': 0.0558094, 'j': 0.0009033, 'k': 0.0050529, 'l': 0.0331490, 'm': 0.0202124, 'n': 0.0564513,
	'o': 0.0596302, 'p': 0.0137645, 'q': 0.0008606, 'r': 0.0497563, 's': 0.0515760, 't': 0.0729357, 'u': 0.0225134,
	'v': 0.0082903, 'w': 0.0171272, 'x': 0.0013692, 'y': 0.0145984, 'z': 0.0007836, ' ': 0.1918182}

// ScoreSorter sorts results by score
type ScoreSorter []Result

func (a ScoreSorter) Len() int           { return len(a) }
func (a ScoreSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ScoreSorter) Less(i, j int) bool { return a[i].Score > a[j].Score }

// Result key attempted, score and result
type Result struct {
	Key       byte
	Score     float64
	Plaintext []byte
}

func countBytes(data []byte) map[byte]int {
	res := make(map[byte]int)
	for _, v := range data {
		if res[v] != 0 {
			res[v]++
		} else {
			res[v] = 1
		}
	}
	return res
}

func getEnglishScore(data []byte) float64 {
	var score float64
	totalChars := len(data)

	data = bytes.ToLower(data)
	counts := countBytes(data)

	for b, c := range counts {
		score += math.Sqrt(FrequencyTable[b] * float64(c) / float64(totalChars))
	}

	return score
}

func singleByteXOR(data []byte, key byte) []byte {
	res := make([]byte, len(data))

	for i := 0; i < len(data); i++ {
		res[i] = data[i] ^ key
	}

	return res
}

func singleByteXORBruteForce(data []byte) {
	var results []Result

	for i := byte(0); ; i++ {
		plaintextCandidate := singleByteXOR(data, i)
		candidateScore := getEnglishScore(plaintextCandidate)

		res := Result{
			Key:       i,
			Score:     candidateScore,
			Plaintext: plaintextCandidate}

		results = append(results, res)
		if i == 255 {
			break
		}
	}

	sort.Sort(ScoreSorter(results))
	fmt.Println("Key:", results[0].Key)
	fmt.Println("Plaintext:", string(results[0].Plaintext))
	fmt.Println("Score:", results[0].Score)
}

func main() {
	data, err := hex.DecodeString("4a59504646515659")
	if err != nil {
		log.Fatal(err)
	}

	singleByteXORBruteForce(data)

}
