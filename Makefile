build:
	go build -o ./build/proxy/out proxy-server/cmd/main.go 		&& \
	go build -o ./build/webapi/out web-api/cmd/main.go