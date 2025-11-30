run:
	go run main.go

lint:
	golangci-lint run ./...

test:
	go test -v ./...

tidy:
	go mod tidy

pipeline-test:
	act push