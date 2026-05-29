.PHONY: build build-linux build-windows clean run

BINARY_NAME=sf

build:
	go build -o $(BINARY_NAME) .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe .

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(BINARY_NAME).exe

run:
	go run . -config=config.json
