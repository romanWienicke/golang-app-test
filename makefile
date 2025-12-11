run:
	go run main.go

lint:
	golangci-lint run ./...

test:
	go test --count=1 -v ./...

app-test:
	go test -v --count=1 ./app/...

service-test:
	go test -v --count=1 ./service/...

tidy:
	go mod tidy

pipeline-test:
	act push

fpush:
	git push --force-with-lease

commit-ammend:
	git commit --amend --no-edit