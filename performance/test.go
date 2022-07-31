package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/emirpasic/gods/trees/btree"
)

func create_SHA256_B64(data string) (b64hash string) {

	hash := sha256.New()
	// Append the new hash based on the previous hash + new payload
	hash.Write([]byte(string(data)))
	b64hash = base64.RawStdEncoding.EncodeToString(hash.Sum(nil))

	return b64hash

}

func main() {

	// Test the specified number of iterations
	const benchmarkLoops = 1_000_000

	m := make(map[string]string)

	// Benchmark standard GO map
	for i := 0; i < benchmarkLoops; i++ {

		m[create_SHA256_B64(string(i))] = string(i)

	}

	// Benchmark btree
	tree := btree.NewWithStringComparator(8) // empty (keys are of type string)

	for i := 0; i < benchmarkLoops; i++ {
		tree.Put(create_SHA256_B64(string(i)), string(i))
	}

	//fmt.Println(tree)
	//spew.Dump(tree.Root.Children)
	fmt.Println(tree.Get("zOa9x8ae3I2k6KB/xHjUhmX4bkN58fsy+t8B2Eszznc"))

}
