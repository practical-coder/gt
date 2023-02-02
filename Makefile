.PHONY:fmt test scheck vet
fmt:
	go fmt ./...
scheck:
	staticcheck ./...
test:
	go test -v ./...
vet: fmt
	go vet ./...