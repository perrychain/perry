PROJECT_NAME_CPU := perry-verify-cpu
PROJECT_NAME_GPU := perry-verify-gpu
current_dir = $(shell pwd)

all: build_cpu verify libed25519

build_cpu:
	@echo "\n....Building $(PROJECT_NAME)"
# gcc main.c cJSON.c lib/base64/*.c lib/ed25519/src/*.c -lsodium -o $(PROJECT_NAME)
	gcc -Wall -O2 main.c lib/cJSON/cJSON.c  -lsodium -o bin/$(PROJECT_NAME_CPU)

verify:
	@echo "\n....libsodium benchmarks"
	gcc -Wall -O2 verify-libsodium.c -o bin/verify-libsodium -lsodium

libed25519:
	@echo "\n....libed25519 benchmarks"
	gcc -Wall -O2 lib/ed25519/test.c lib/ed25519/src/*.c -o bin/verify-libed25519

benchmark: verify libed25519
	./bin/verify-libsodium > ./benchmarks/libsodium.txt
	./bin/verify-libed25519 > ./benchmarks/libed25519.txt
	go test -bench=. -benchtime=10000x $(current_dir)/../pkg/wallet/wallet_test.go > $(current_dir)/benchmarks/go-crypto-ed25519.txt

benchmark_go_map_btree: 
	go test -bench=. -benchtime=1000000x -benchmem performance_test.go > $(current_dir)/benchmarks/benchmark_go_map_btree.txt





