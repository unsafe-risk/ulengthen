build_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/ulengthen_linux_amd64 .
