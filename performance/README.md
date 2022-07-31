# Perry performance enhancements

Repository for Perry blockchain verification improvements in C (CPU), CUDA (GPU) and feature support for AVX512 extensions (CPU) . Used for validating message signatures and blocks from Perry nodes and local blockchain DB on disk.

## Pre-req

For CPU benchmarks the libsodium library must be pre-installed, http://libsodium.org

Ubuntu:

```
sudo apt-get install libsodium-dev
```

Mac:

```
brew install libsodium
```

For GPU confirm the [CUDA/NVIDIA libraries are correctly installed](https://docs.nvidia.com/cuda/cuda-installation-guide-linux/index.html) as per your system requirements.

## Build

CPU benchmarks:

```
make build_cpu
./bin/perry-verify-cpu ~/.perry/blockchain-db.json
```

GPU benchmarks (TBC):

```
make build_gpu
./bin/perry-verify-gpu ~/.perry/blockchain-db.json
```

## GO map vs B-tree 

Benchmark a standard GO map which contains N unique elements (SHA256 sum, base64 encoded) and compare the differences using two different B-tree implementations [github.com/tidwall/btree](https://github.com/tidwall/btree) and [github.com/emirpasic/gods/trees/btree](https://github.com/emirpasic/gods/trees/btree) for inserting and fetching N unique elements.

Future scope for fetching a range of users messages from a specified public-key using a map or b-tree implementation within Perry.

```
make benchmark_go_map_btree
...;
go test -bench=. -benchtime=1000000x performance_test.go > benchmarks/benchmark_go_map_btree.txt
```

```
goos: darwin
goarch: arm64
Benchmark_RandomArray-8            	 1000000	     11244 ns/op	     450 B/op	      12 allocs/op
Benchmark_MapCreation-8            	 1000000	       424.1 ns/op	     259 B/op	       6 allocs/op
Benchmark_MapRandomAccess-8        	 1000000	       107.7 ns/op	       0 B/op	       0 allocs/op
Benchmark_TD_BtreeCreation-8       	 1000000	      1201 ns/op	     321 B/op	       7 allocs/op
Benchmark_TD_BtreeRandomAccess-8   	 1000000	       858.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_TreeCreation-8           	 1000000	      2114 ns/op	     326 B/op	      11 allocs/op
Benchmark_TreeRandomAccess-8       	 1000000	      1943 ns/op	      16 B/op	       1 allocs/op
PASS
ok  	command-line-arguments	21.802s
```

