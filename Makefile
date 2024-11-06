.PHONY: vendor
vendor: 
	go mod tidy && go mod vendor

.PHONY: server/start
server/start:
	docker-compose -f docker-compose.yml up -d  --remove-orphans

.PHONY: server/stop
server/stop:
	docker-compose -f docker-compose.yml down


GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

.PHONY: build/app
build/app:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o build/app cmd/main.go