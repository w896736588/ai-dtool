# Step 1: Detect operating system (compatible with Windows/Linux)
# Windows has built-in OS=Windows_NT environment variable, Linux does not
ifeq ($(OS),Windows_NT)
    CURRENT_OS := windows
    ENV_CMD := set
else
    CURRENT_OS := linux
    ENV_CMD := export
endif

# Define common variables for easy maintenance
GO_BUILD_FLAGS := -ldflags "-s -w"
OUTPUT_PATH := ./build/dtool.exe
MAIN_GO_PATH := ./cmd/dtool/main.go

# Phony targets to avoid conflicts with files of the same name
.PHONY: dev_tool_windows dev_tool_linux make_all clean

# Compile target for Windows (fixed typo: widows -> windows)
dev_tool_windows:
	@echo "========== Current runtime environment: $(CURRENT_OS) =========="
	@echo "Setting Go build environment variables..."
	# Set temporary variables based on system (only effective for this build)
	$(ENV_CMD) CGO_ENABLED=1 && $(ENV_CMD) GOOS=windows && $(ENV_CMD) GOARCH=amd64
	@echo "Executing go mod tidy..."
	go mod tidy
	@echo "Building Go program to $(OUTPUT_PATH)..."
	# Specify environment variables directly when building (highest priority)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(OUTPUT_PATH) $(MAIN_GO_PATH)
	@echo "Checking if build product is tracked by git..."
	git ls-files --stage $(OUTPUT_PATH) || echo "Note: $(OUTPUT_PATH) is not tracked by git (not an error)"
	@echo "========== Build completed: $(OUTPUT_PATH) =========="

# Compile target for Linux (reuse Windows logic)
dev_tool_linux: dev_tool_windows

# Main target: auto-adapt to system and compile
make_all:
	@if [ "$(CURRENT_OS)" = "windows" ]; then \
		make dev_tool_windows; \
	else \
		make dev_tool_linux; \
	fi

# Clean build products (optional)
clean:
	@echo "Cleaning $(OUTPUT_PATH)..."
	@if [ -f $(OUTPUT_PATH) ]; then \
		rm -f $(OUTPUT_PATH); \
		echo "Clean completed"; \
	else \
		echo "No build products to clean"; \
	fi