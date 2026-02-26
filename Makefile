
hello:
	echo "xxx"

build:
	CGO_CFLAGS="-Wno-return-local-address" go build -o $(or $(OUTPUT),./w7panel-offline) .

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_CFLAGS="-Wno-return-local-address" go build -o $(or $(OUTPUT),./w7panel-offline) .

dev:
	CGO_CFLAGS="-Wno-return-local-address" go run main.go server:start

test:
	go test ./...
