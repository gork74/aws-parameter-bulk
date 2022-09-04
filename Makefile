export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: clean
clean:
	rm -rf ./bin
	go clean

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: test
test:
	go test ./... -parallel=1 -cover -coverprofile cover.out | sed ''/PASS/s//$(shell printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$(shell printf "\033[31mFAIL\033[0m")/''

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/aws-parameter-bulk-darwin-amd64 main/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/aws-parameter-bulk-darwin-arm64 main/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/aws-parameter-bulk-linux-amd64 main/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/aws-parameter-bulk-linux-arm64 main/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -o ./bin/aws-parameter-bulk-linux-armhf main/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/aws-parameter-bulk-windows-amd64.exe main/main.go

.PHONY: dev
dev:
	SSM_LOG_LEVEL=debug go run main/main.go

# install goreleaser first
.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: setversion
setversion:
	if [ -z "$(GITHUB_REF_NAME)" ]; then echo "GITHUB_REF_NAME is not set"; exit 1; fi
	sed "s/v[0-9]\.[0-9]\.[0-9]/${GITHUB_REF_NAME}/g" conf/version.go > conf/version.temp
	rm conf/version.go
	mv conf/version.temp conf/version.go
