# ==========================================
# NVR Project Makefile
# ==========================================

# Docker Build Name
DOCKER_IMAGE_NAME = vmsnvr
VERSION = 0.1

# -- Output Binaries --
# The C++ worker binary name (must match what main.go looks for)
WORKER_BIN_NAME := nvr_worker
# The Go manager binary name
SERVICE_BIN_NAME := nvr_service

# -- Directories --
GO_DIR := nvr_core
CPP_ENGINE_DIR := cpp_engine
CPP_ENGINE_BUILD_DIR := $(CPP_ENGINE_DIR)/build

# -- Tools --
GO := go
CMAKE := cmake
MKDIR := mkdir -p
RM := rm -rf
CP := cp

# -- Targets --
.PHONY: all clean build-cpp build-go docker-build docker-run

# Default target: Build everything
all: build-go build-cpp

# 1. Build C++ Worker
# Steps: Create build dir -> Run CMake -> Run Make -> Copy binary to root
build-cpp:
	@echo "--- Building C++ Worker ---"
	$(MKDIR) $(CPP_ENGINE_BUILD_DIR)
	cd $(CPP_ENGINE_BUILD_DIR) && $(CMAKE) .. && $(MAKE)
	# Copy the compiled binary from build folder to project root
	$(CP) $(CPP_ENGINE_BUILD_DIR)/nvr_worker ./$(WORKER_BIN_NAME)
	@echo "✔ C++ Worker built successfully: ./$(WORKER_BIN_NAME)"

# 2. Build Go NVR Service
build-go:
	@echo "--- Building Go NVR Service ---"
	# Ensure go.mod exists (create if missing)
	@[ -f go.mod ] || $(GO) mod init nvr-core
	$(GO) mod tidy
	cd $(GO_DIR) && $(GO) build -o $(SERVICE_BIN_NAME)
	@echo "✔ Go Manager built successfully: ./$(SERVICE_BIN_NAME)"

# 3. Clean Build Artifacts
clean:
	@echo "--- Cleaning ---"
	$(RM) $(CPP_ENGINE_BUILD_DIR)
	$(RM) $(WORKER_BIN_NAME)
	$(RM) $(SERVICE_BIN_NAME)
	@echo "✔ Clean complete"

# Build Docker image
docker:
	docker build --platform linux/amd64 -t $(DOCKER_IMAGE_NAME) .
# 	docker build -t $(DOCKER_IMAGE_NAME) .

dockersave:
	docker save $(DOCKER_IMAGE_NAME):latest | gzip > ../nvr_image.$(VERSION).tar.gz

export:
	docker create --name nvr-extractor $(DOCKER_IMAGE_NAME)
	docker cp nvr-extractor:/app/nvr_service ./dist/nvr_service
	docker cp nvr-extractor:/app/nvr_worker ./dist/nvr_worker
	docker rm nvr-extractor

# Run the Docker container
docker-run:
	docker run -it --rm $(DOCKER_IMAGE_NAME)
