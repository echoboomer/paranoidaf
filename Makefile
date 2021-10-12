build:
	echo "Compiling for compatible platforms"
	GOOS=darwin GOARCH=amd64 go build -o bin/paranoidaf
