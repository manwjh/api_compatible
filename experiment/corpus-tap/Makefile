.PHONY: build docker run

build:
	go build -o bin/corpus-tap ./cmd/corpus-tap

docker:
	docker build -t corpus-tap:dev .

run: build
	./bin/corpus-tap
