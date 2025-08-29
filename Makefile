.PHONY: build install uninstall clean run dev deps

BINARY_NAME=go-ranger
INSTALL_DIR=$$HOME/.local/bin
GOBUILD=go build -ldflags="-s -w" -trimpath

# Colors
GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m

build:
	@echo "${BLUE}Building ${BINARY_NAME}...${NC}"
	${GOBUILD} -o ${BINARY_NAME} .
	@echo "${GREEN}Build complete: ./${BINARY_NAME}${NC}"

install: build
	@echo "${BLUE}Installing ${BINARY_NAME} to ${INSTALL_DIR}...${NC}"
	@mkdir -p ${INSTALL_DIR}
	@cp ${BINARY_NAME} ${INSTALL_DIR}/
	@chmod +x ${INSTALL_DIR}/${BINARY_NAME}
	@echo "${GREEN}Installation complete!${NC}"
	@echo "Add to PATH: export PATH=\$$HOME/.local/bin:\$$PATH"

uninstall:
	@echo "${BLUE}Removing ${BINARY_NAME}...${NC}"
	@rm -f ${INSTALL_DIR}/${BINARY_NAME}
	@echo "${GREEN}Uninstall complete!${NC}"

run: build
	@echo "${BLUE}Running ${BINARY_NAME}...${NC}"
	./${BINARY_NAME}

dev:
	@echo "${BLUE}Starting development mode...${NC}"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Running with go run..."; \
		go run .; \
	fi

clean:
	@echo "${BLUE}Cleaning up...${NC}"
	@rm -f ${BINARY_NAME}
	@echo "${GREEN}Clean complete${NC}"

deps:
	@echo "${BLUE}Downloading dependencies...${NC}"
	go mod download
	@echo "${GREEN}Dependencies updated${NC}"

help:
	@echo "Available commands:"
	@echo "  make build    - Build binary"
	@echo "  make install  - Install to ~/.local/bin"
	@echo "  make run      - Build and run"
	@echo "  make dev      - Run with hot reload"
	@echo "  make clean    - Clean build files"
	@echo "  make deps     - Download dependencies"