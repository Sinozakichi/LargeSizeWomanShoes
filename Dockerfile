ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

# 安裝 CA 憑證並更新
RUN apt update && apt install -y ca-certificates && update-ca-certificates

RUN go build -v -o /run-app .

# 建置全新的容器
FROM debian:bookworm

# 設置環境變數
# ENV GO_ENV=release

# 安裝 CA 憑證，確保它在最終映像中也能存在。本地端運行正常通常是因為本地的操作系統會處理 CA 憑證，並且網路配置是開放的。但在 fly.io 上，容器可能缺少信任根憑證，所以要手動於容器中安裝最新的 CA 憑證
RUN apt update && apt install -y ca-certificates

WORKDIR /app
COPY --from=builder /run-app /usr/local/bin/
# 複製靜態資源檔案 (js 和 html)
COPY statics /app/static
COPY scripts /app/script
CMD ["run-app"]
