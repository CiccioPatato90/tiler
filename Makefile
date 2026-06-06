BINARY := bin/tiler

.PHONY: all build run clean

all: build

build:
	mkdir -p bin
	# remove when a real git repo is attached
	go build -buildvcs=false -o $(BINARY) .

run: build
	./$(BINARY)

clean:
	rm -rf bin
