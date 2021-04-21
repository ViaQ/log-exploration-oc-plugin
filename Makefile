CONTAINER_ENGINE?=podman
EXECUTABLE:=log-exploration-oc-plugin
PACKAGE:=github.com/ViaQ/log-exploration-oc-plugin
IMAGE_PUSH_REGISTRY:=quay.io/emishra/$(EXECUTABLE)
VERSION:=${shell git describe --tags --always}
BUILDTIME := ${shell date -u '+%Y-%m-%d_%H:%M:%S'}
BUILD_DIR:=./bin

.PHONY: build test clean image image-publish
build: test
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build cmd/oc-historical_logs.go
	chmod +x oc-historical_logs
	sudo mv oc-historical_logs /usr/local/bin/.

test:
	go test ./pkg/... -cover

test-cover:
	go test ./pkg/... -coverprofile=coverage.out && go tool cover -html=coverage.out

clean:
	rm -rf $(BUILD_DIR)/

image: build
	$(CONTAINER_ENGINE) build . -t ${IMAGE_PUSH_REGISTRY}:${VERSION}

image-publish: image
	$(CONTAINER_ENGINE) push ${IMAGE_PUSH_REGISTRY}:${VERSION}

