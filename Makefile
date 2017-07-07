CGO_ENABLED=0

all: windows linux darwin

linux:
	GOOS=linux GOARCH=amd64 go build -o steamroll_linux cmd/steamroll/*.go

windows:
	GOOS=windows GOARCH=amd64 go build -o steamroll.exe cmd/steamroll/*.go

darwin:
	GOOS=darwin GOARCH=amd64 go build -o steamroll_darwin cmd/steamroll/*.go

clean:
	rm steamroll_linux
	rm steamroll.exe
	rm steamroll_darwin
