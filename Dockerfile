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
ENV GO_ENV=release

# 更新系統並安裝必要的依賴項
RUN apt update && apt install -y --no-install-recommends \
    # 安裝 CA 憑證
    ca-certificates \
    # 安裝 Chromium 瀏覽器
    chromium \
    dbus-x11 \
    libatk-bridge2.0-0 \
    libgtk-3-0 \
    libgbm1 \
    libasound2 \
    libnss3 \
    libx11-xcb1 \
    fonts-liberation \
    --no-install-recommends \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /run-app /usr/local/bin/
# 複製靜態資源檔案 (css、js 和 html)
COPY css /app/static
COPY statics /app/static
COPY scripts /app/script

# 啟動 dbus-daemon 和你的應用
CMD ["run-app"]
# CMD dbus-daemon --system --fork && run-app
