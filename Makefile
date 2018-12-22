
all:
	mkdir -p dist
	cp -p sh/*.sh dist/
	go build -ldflags="-s -w" -o dist/juun.search control/search.go control/io.go
	go build -ldflags="-s -w" -o dist/juun.service service/*.go
	go build -ldflags="-s -w" -o dist/juun.updown control/updown.go control/io.go
