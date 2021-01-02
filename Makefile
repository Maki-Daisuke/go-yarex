all: *.go cmd/yarexgen/main.go
	cd cmd/yarexgen; go build .

test: all cmd/yarexgen/yarexgen
	go generate
	go test

clean:
	rm cmd/yarexgen/yarexgen
	rm *_yarex_test.go
