package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
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

// PTScoreSorter sorts results by score
type PTScoreSorter []PlaintextResult

func (a PTScoreSorter) Len() int           { return len(a) }
func (a PTScoreSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PTScoreSorter) Less(i, j int) bool { return a[i].Score > a[j].Score }

// Result key attempted, score and result
type Result struct {
	Key       byte
	Score     float64
	Plaintext []byte
}

// PlaintextResult key attempted, score and result
type PlaintextResult struct {
	Key       []byte
	Score     float64
	Plaintext []byte
}
type keySizeDistance struct {
	KeySize  int
	Distance float64
}

type distanceSorter []keySizeDistance

func (a distanceSorter) Len() int           { return len(a) }
func (a distanceSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a distanceSorter) Less(i, j int) bool { return a[i].Distance < a[j].Distance }

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

func multiByteXOR(data, key []byte) []byte {
	res := make([]byte, len(data))

	for i := 0; i < len(data); i++ {
		res[i] = data[i] ^ key[i%len(key)]
	}

	return res
}

func singleByteXORBruteForce(data []byte) byte {
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
	return results[0].Key
}

// Computes the edit distance/Hamming distance between two equal-length strings
func hammingDistance(s1, s2 []byte) float64 {
	var distance float64

	if len(s1) != len(s2) {
		log.Fatal("Requires equal-length strings")
	}

	for i := 0; i < len(s1); i++ {
		diff := strconv.FormatInt(int64(s1[i]^s2[i]), 2)
		for _, b := range []byte(diff) {
			if string(b) == "1" {
				distance++
			}
		}
	}

	return distance
}

func breakRepeatingKeyXOR(data []byte) {
	var distance float64
	var normalizedDistance float64
	var normalizedDistances []keySizeDistance

	// For each key size
	for keySize := 2; keySize <= 42; keySize++ {
		// Take the first four key_size worth of bytes
		chunks := make([][]byte, 4)
		c := 0
		for i := 0; i <= len(data); i += keySize {
			chunks[c] = make([]byte, keySize)
			chunks[c] = data[i : i+keySize]
			c++
			if c == 4 {
				break
			}
		}

		// Sum differances between each pair of chunks
		distance = 0
		distance += hammingDistance(chunks[0], chunks[1])
		distance += hammingDistance(chunks[0], chunks[2])
		distance += hammingDistance(chunks[0], chunks[3])
		distance += hammingDistance(chunks[1], chunks[2])
		distance += hammingDistance(chunks[1], chunks[3])
		distance += hammingDistance(chunks[2], chunks[3])

		// Take the average
		distance /= float64(6)

		// Normalize the result by dividing by key_size and store it by the key size
		normalizedDistance = distance / float64(keySize)
		normalizedDistances = append(normalizedDistances, keySizeDistance{keySize, normalizedDistance})
	}

	// The key_sizes with the smallest normalized edit distances are the most likely ones
	sort.Sort(distanceSorter(normalizedDistances))
	var possibleKeySizes []int
	for _, v := range normalizedDistances[0:3] {
		possibleKeySizes = append(possibleKeySizes, v.KeySize)
	}

	var plaintextResults []PlaintextResult

	// Now we can try which one is really the correct one among the top 3 most likely sizes
	for _, keySize := range possibleKeySizes {
		var key []byte

		// Break the ciphertext into blocks of key_size length
		for i := 0; i < keySize; i++ {
			var block []byte

			// Transpose the blocks: make a block that is the i-th byte of every block
			for j := i; j < len(data); j += keySize {
				block = append(block, data[j])
			}

			key = append(key, singleByteXORBruteForce(block))
		}
		plaintextResult := PlaintextResult{}
		plaintextResult.Plaintext = multiByteXOR(data, key)
		plaintextResult.Key = key
		plaintextResult.Score = getEnglishScore(plaintextResult.Plaintext)

		plaintextResults = append(plaintextResults, plaintextResult)
	}

	sort.Sort(PTScoreSorter(plaintextResults))
	fmt.Println(plaintextResults[0].Key)
	fmt.Println(string(plaintextResults[0].Key))
	fmt.Println(plaintextResults[0].Score)
	fmt.Println(string(plaintextResults[0].Plaintext))
}

func main() {
	input, err := os.Open("6.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	decoder := base64.NewDecoder(base64.StdEncoding, input)

	data, err := ioutil.ReadAll(decoder)
	if err != nil {
		log.Fatal(err)
	}

	breakRepeatingKeyXOR(data)
}
