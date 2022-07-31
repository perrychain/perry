package wallet_test

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/emirpasic/gods/trees/btree"
	tidwall_btree "github.com/tidwall/btree"
)

var randomArray []string
var m map[string]string
var tree *btree.Tree

var td_btree tidwall_btree.Map[string, string]

func create_SHA256_B64(data string) (b64hash string) {

	hash := sha256.New()
	// Append the new hash based on the previous hash + new payload
	hash.Write([]byte(string(data)))
	b64hash = base64.RawStdEncoding.EncodeToString(hash.Sum(nil))

	return b64hash

}

func generate_random_key(keyrange int) (randomKey, expectedValue string) {

	// Random seed required each time
	rand.Seed(time.Now().UnixNano())

	randomNum := rand.Intn(keyrange)
	randomKey = create_SHA256_B64(fmt.Sprintf("HELLO_WORLD_%d", randomNum))

	// Check the value is as expected
	expectedValue = fmt.Sprintf("DB_INDEX_%d", randomNum)

	return

}

func Benchmark_RandomArray(b *testing.B) {

	randomArray = make([]string, b.N)

	lm := make(map[string]bool)

	for n := 0; n < b.N; n++ {
		lm[create_SHA256_B64(fmt.Sprintf("HELLO_WORLD_%d", n))] = true
	}

	// Add random elements
	for n := 0; n < b.N; n++ {
		randomKey, _ := generate_random_key(b.N)
		randomArray[n] = randomKey
	}

}

func Benchmark_MapCreation(b *testing.B) {

	m = make(map[string]string, b.N)

	for n := 0; n < b.N; n++ {
		m[create_SHA256_B64(fmt.Sprintf("HELLO_WORLD_%d", n))] = fmt.Sprintf("DB_INDEX_%d", n)
	}

}

func Benchmark_MapRandomAccess(b *testing.B) {

	for n := 0; n < len(randomArray); n++ {
		key := randomArray[n]

		if key == "" {
			fmt.Println("Key not found")
		} else {

			// Fetch from the map
			val := m[key]

			if val == "" {
				fmt.Println("Val not found")
			}

		}

	}

}

func Benchmark_TD_BtreeCreation(b *testing.B) {

	for n := 0; n < b.N; n++ {
		td_btree.Set(create_SHA256_B64(fmt.Sprintf("HELLO_WORLD_%d", n)), fmt.Sprintf("DB_INDEX_%d", n))
	}

}

func Benchmark_TD_BtreeRandomAccess(b *testing.B) {

	for n := 0; n < len(randomArray); n++ {
		key := randomArray[n]

		if key == "" {
			fmt.Println("Key not found")
		} else {

			// Fetch from the map
			val, _ := td_btree.Get(key)

			if val == "" {
				fmt.Println("Val not found")
			}

		}

	}

}

func Benchmark_TreeCreation(b *testing.B) {

	// Benchmark btree
	tree = btree.NewWithStringComparator(8) // empty (keys are of type string)

	for n := 0; n < b.N; n++ {
		tree.Put(create_SHA256_B64(fmt.Sprintf("HELLO_WORLD_%d", n)), fmt.Sprintf("DB_INDEX_%d", n))
	}

}

func Benchmark_TreeRandomAccess(b *testing.B) {

	for n := 0; n < len(randomArray); n++ {
		key := randomArray[n]

		if key == "" {
			fmt.Println("Key not found")
		} else {

			// Fetch from the map
			val, _ := tree.Get(key)

			if val == "" {
				fmt.Println("Val not found")
			}

		}

	}

}
