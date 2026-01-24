.PHONY: setup fmt tidy lint test run build clean

# 初始化项目依赖
setup:
	@go mod download
	@go mod tidy

# 格式化代码
fmt:
	@gofmt -l -w .
	@go mod tidy

# 整理依赖
tidy:
	@go mod tidy -v

# 运行 lint
lint:
	@golangci-lint run ./...

# 运行测试
test:
	@go test -race -shuffle=on -short -failfast ./...

# 运行服务器
run:
	@go run cmd/server/main.go

# 使用配置文件运行
run-config:
	@go run cmd/server/main.go --config=./config/config.yaml

# 编译二进制文件
build:
	@go build -o bin/ai-gateway cmd/server/main.go

# 清理构建产物
clean:
	@rm -rf bin/

# 生成代码（如果使用 go generate）
gen:
	@go generate ./...

# Docker 构建
docker-build:
	docker build -t ai-gateway:latest .

# Docker 运行 (使用 compose)
docker-up:
	docker-compose up -d

# Docker 停止
docker-down:
	docker-compose down


# 帮助
help:
	@echo "可用目标："
	@echo "  setup      - 下载并整理依赖"
	@echo "  fmt        - 格式化代码"
	@echo "  tidy       - 整理 go.mod"
	@echo "  lint       - 运行 lint"
	@echo "  test       - 运行测试"
	@echo "  run        - 使用默认配置运行服务器"
	@echo "  run-config - 使用配置文件运行服务器"
	@echo "  build      - 编译二进制文件"
	@echo "  clean      - 清理构建产物"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-up    - 启动 Docker 服务"
	@echo "  docker-down  - 停止 Docker 服务"