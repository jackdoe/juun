GOARCH ?= "amd64"
GOOS ?= "darwin"
all:
	mkdir -p dist
	cp -p sh/*.sh dist/
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.search control/search.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.import control/import.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.service service/*.go
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags="-s -w" -o dist/juun.updown control/updown.go
clean:
	rm dist/juun.* dist/*.sh
