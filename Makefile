
all:
	go build -ldflags="-s -w" -o juun.search search.go
	go build -ldflags="-s -w" -o juun.service service.go  query.go
	go build -ldflags="-s -w" -o juun.updown updown.go
