EXECUTABLE:=log-exploration-oc-plugin
PACKAGE:=github.com/ViaQ/log-exploration-oc-plugin
IMAGE_PUSH_REGISTRY:=docker://quay.io/openshift-logging/$(EXECUTABLE)
VERSION:=${shell git describe --tags --always}
BUILDTIME := ${shell date -u '+%Y-%m-%d_%H:%M:%S'}
LDFLAGS:= -s -w -X '${PACKAGE}/pkg/version.Version=${VERSION}' \
					-X '${PACKAGE}/pkg/version.BuildTime=${BUILDTIME}'
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
	docker build . -t ${EXECUTABLE}:${VERSION}

image-publish: image
	docker push ${EXECUTABLE}:${VERSION} ${IMAGE_PUSH_REGISTRY}:${VERSION}

