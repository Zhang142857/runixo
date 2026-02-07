.PHONY: build run clean

BINARY=runixo-server

build:
	go build -o $(BINARY) ./cmd/server/

run: build
	./$(BINARY) config.yaml

clean:
	rm -f $(BINARY)
