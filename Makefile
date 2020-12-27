all: *.go cmd/yarexgen/main.go
	cd cmd/yarexgen; go build .

test: all cmd/yarexgen/yarexgen
	go generate
	go test
