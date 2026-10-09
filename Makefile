# Название выходного файла и папка для сборок
APP_NAME = wireproxy
BUILD_DIR = build
MAIN_PKG = ./cmd/wireproxy

# Флаги линковщика: -s и -w вырезают отладочную информацию (debug symbols), 
# что сильно уменьшает размер итоговых бинарников.
# Версия вшивается в бинарник (wireproxy --version)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all clean build build-all linux windows darwin android-bin aar help test

# Действие по умолчанию (просто make)
all: clean build

help:
	@echo "Доступные команды:"
	@echo "  make             - собрать под текущую ОС (по умолчанию)"
	@echo "  make build-all   - собрать бинарники под все популярные ОС (Linux, Windows, macOS)"
	@echo "  make linux       - собрать только для Linux (amd64, arm64)"
	@echo "  make windows     - собрать только для Windows (amd64, arm64)"
	@echo "  make darwin      - собрать только для macOS (amd64, arm64)"
	@echo "  make android-bin - собрать консольный бинарник для Android (arm64)"
	@echo "  make aar         - собрать Android библиотеку (.aar) через gomobile"
	@echo "  make test        - go vet + go test -race (включая интеграционные тесты туннеля)"
	@echo "  make clean       - удалить папку $(BUILD_DIR)"

test:
	go vet ./...
	go test -race -count=1 ./...

build:
	@echo "==> Сборка для текущей ОС..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)

linux:
	@echo "==> Сборка для Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PKG)

windows:
	@echo "==> Сборка для Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe $(MAIN_PKG)

darwin:
	@echo "==> Сборка для macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)

android-bin:
	@echo "==> Сборка консольного бинарника для Android (Termux / Root)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=android GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-android-arm64 $(MAIN_PKG)

# ВАЖНО: Для выполнения 'make aar' у тебя должен быть установлен gomobile,
# Android SDK и NDK. Пакет ./mobile должен существовать (обсуждали ранее).
aar:
	@echo "==> Сборка Android библиотеки (.aar)..."
	@mkdir -p $(BUILD_DIR)
	gomobile bind -target=android -o $(BUILD_DIR)/$(APP_NAME).aar -ldflags="-s -w" ./mobile

build-all: linux windows darwin android-bin
	@echo "==> Успешно! Все бинарники собраны в папке $(BUILD_DIR)/"

clean:
	@echo "==> Очистка папки сборки..."
	rm -rf $(BUILD_DIR)