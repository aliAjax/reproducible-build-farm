APP=buildfarm
.PHONY: fmt test vet build run smoke docker
fmt:
	gofmt -w $$(find . -name '*.go')
test:
	go test ./...
vet:
	go vet ./...
build:
	go build ./...
run:
	go run ./cmd/buildfarm
smoke:
	bash scripts/smoke.sh
docker:
	docker build -t $(APP):dev .
