# Build:
BINARY_DIR=bin
BINARY_NAME=master
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
RELEASE_NAME=tinykv

# Test:
COVERAGE_PROFILE=cover.out

# Docker dev:
DOCKER_IMAGE=tinykv-master
DOCKER_IMAGE_DEV=$(DOCKER_IMAGE)-dev
DOCKER_IMAGE_WATCH=$(DOCKER_IMAGE)-watch
WATCH_MASTER_PORT=3000
WATCH_VOL1_PORT=3001
WATCH_VOL2_PORT=3002
WATCH_VOL3_PORT=3003
WATCH_MASTER_INDEX_PATH=tmp/indexdb
WATCH_VOL1_PATH=tmp/vol1
WATCH_VOL2_PATH=tmp/vol2
WATCH_VOL3_PATH=tmp/vol3

# Docker config:
HOST=localhost
REPLICAS=3
VOLUME=3

.PHONY: vendor

all: format lint coverage build

# Build:
setup:
	@go mod download

build:
	@go build -buildvcs=false -o $(BINARY_PATH) ./cmd/master

clean:
	@go clean --cache
	@rm -rf $(BINARY_DIR)
	@rm $(COVERAGE_PROFILE)

vendor:
	@GO111MODULE=on go mod vendor

watch:
	$(eval PACKAGE_NAME=$(shell head -n 1 go.mod | cut -d ' ' -f2))
	@docker build -t $(DOCKER_IMAGE_WATCH) --target watch .
	@docker run -it --rm \
		-w /go/src/$(PACKAGE_NAME) \
		-v $(shell pwd):/go/src/$(PACKAGE_NAME) \
		-p $(WATCH_MASTER_PORT):$(WATCH_MASTER_PORT) \
		-p $(WATCH_VOL1_PORT):$(WATCH_VOL1_PORT) \
		-p $(WATCH_VOL2_PORT):$(WATCH_VOL2_PORT) \
		-p $(WATCH_VOL3_PORT):$(WATCH_VOL3_PORT) \
		--name $(DOCKER_IMAGE_WATCH) \
		$(DOCKER_IMAGE_WATCH) \
		--build.cmd "go build -race -buildvcs=false -o $(BINARY_PATH) ./cmd/master && \
					 ./volume/kill_all.sh && \
					 rm -r tmp || true && \
					 PORT=$(WATCH_VOL1_PORT) VOLUME=$(WATCH_VOL1_PATH) ./volume/setup.sh && \
					 PORT=$(WATCH_VOL2_PORT) VOLUME=$(WATCH_VOL2_PATH) ./volume/setup.sh && \
					 PORT=$(WATCH_VOL3_PORT) VOLUME=$(WATCH_VOL3_PATH) ./volume/setup.sh" \
		--build.bin "./$(BINARY_PATH) -db $(WATCH_MASTER_INDEX_PATH) -p $(WATCH_MASTER_PORT) -volumes localhost:$(WATCH_VOL1_PORT),localhost:$(WATCH_VOL2_PORT),localhost:$(WATCH_VOL3_PORT)"

# Test:
test:
	@go test -v -race ./...

coverage:
	@go test -v -race -cover -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func $(COVERAGE_PROFILE)

# Benchmark:
bench:
	@go test -bench=. -run=^a -benchtime=5x ./...

# Format
format:
	@gofmt -s -w .

# Lint
lint:
	@golangci-lint run

# Release:
release:
	@go build -buildvcs=false -o $(BINARY_NAME) ./cmd/master
	@tar -czvf $(RELEASE_NAME).tar.gz $(BINARY_NAME) volume README.md LICENSE
	@rm $(BINARY_NAME)

# Docker dev:
docker-dev-setup:
	@docker build -t $(DOCKER_IMAGE_DEV) --target dev .
	@docker run --rm --name $(DOCKER_IMAGE_DEV) -d -v $(shell pwd):/app $(DOCKER_IMAGE_DEV)

docker-dev-stop:
	@docker stop $(DOCKER_IMAGE_DEV)

docker-check:
	@docker exec $(DOCKER_IMAGE_DEV) make

docker-test:
	@docker exec $(DOCKER_IMAGE_DEV) make test

docker-bench:
	@docker exec $(DOCKER_IMAGE_DEV) make bench

docker-up-volume:
	@docker compose -p tinykv -f docker/docker-compose.yml stop volume
	@docker compose -p tinykv -f docker/docker-compose.yml up -d --scale volume=$(VOLUME) volume

docker-up-master:
	@VOLUMES=$$(for i in `seq 1 $(VOLUME)`; do docker compose -p tinykv -f docker/docker-compose.yml port volume 80 --index $$i | cut -d: -f2-; done | sed 's/^/$(HOST):/;s/$$/\n/' | paste -sd "," - | sed 's/,,/,/g;s/,$$//') \
	REPLICAS=$(REPLICAS) \
	docker compose -p tinykv -f docker/docker-compose.yml up -d master

docker-up: docker-up-volume docker-up-master
