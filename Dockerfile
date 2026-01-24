# Build Frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web/admin
COPY web/admin/package*.json ./
RUN npm install
COPY web/admin ./
RUN npm run build

# Build Backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ai-gateway cmd/server/main.go

# Final Stage
FROM alpine:latest
WORKDIR /app

# Install basic dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary
COPY --from=backend-builder /app/ai-gateway .

# Copy frontend static files
COPY --from=frontend-builder /app/web/admin/dist ./web/admin/dist

# Copy config example as default (user needs to mount real config)
COPY config/config.yaml ./config/config.yaml
COPY .env.example .env

EXPOSE 8081

# Default command
CMD ["./ai-gateway", "--config=./config/config.yaml"]
