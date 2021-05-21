PACKAGE:=github.com/ViaQ/log-exploration-oc-plugin
VERSION:=${shell git describe --tags --always}
BUILDTIME := ${shell date -u '+%Y-%m-%d_%H:%M:%S'}
BUILD_DIR:=./bin
LDFLAGS:= -s -w -X '${PACKAGE}/pkg/version.Version=${VERSION}' \
					-X '${PACKAGE}/pkg/version.BuildTime=${BUILDTIME}'

.PHONY: build install test test-e2e
build: test
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o $(BUILD_DIR)/$(EXECUTABLE) cmd/oc-historical_logs.go

install: build
	chmod +x $(BUILD_DIR)/oc-historical_logs
	sudo mv $(BUILD_DIR)/oc-historical_logs /usr/local/bin/.

test:
	go test ./pkg/... -cover

test-cover:
	go test ./pkg/... -coverprofile=coverage.out && go tool cover -html=coverage.out

test-e2e:
	docker-compose up -d
	@sleep 5
	chmod +x test/e2e/populate_indices.sh
	test/e2e/populate_indices.sh
	go test -v test/e2e/*.go
	docker-compose down
