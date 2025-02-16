# LargeSizeWomanShoes

# 👠 大尺碼女鞋爬蟲專案

本專案旨在爬取台灣多家大尺碼女鞋店鋪網站的商品資訊，大尺碼女鞋店鋪的定義為多數品項有提供>=歐碼 41 碼(腳長>=25cm)的店舖者，整理鞋款、尺碼、顏色等資訊，並提供一個基礎的篩選與渲染界面，方便使用者查找適合的鞋款。

## 🚀 目標網站

目前計畫爬取以下 4 家台灣女鞋網站：

- [D+AF](https://www.daf-shoes.com/)
- [Anns](https://www.anns.tw/)
- (待完成)[Amai](https://www.amai.tw/)
- (待完成)[GraceGift](https://www.gracegift.com.tw/)

## ✨ 專案特色

- 📡 **爬取多家鞋店商品資訊**：包括鞋款名稱、圖片、連結、價格、當前有的尺碼、顏色等
- 🔍 **篩選與搜尋功能**：根據尺碼、顏色、品項、跟高、品牌等條件進行篩選
- 🌐 **前端展示**：簡單的 Web 介面，讓使用者可以瀏覽與篩選商品
- 🛠 **技術**：使用 Go 進行爬蟲開發，前端採用 Bootstrap Template

## 🏗 環境需求

請確保你的環境已安裝：

- Go 1.21+

## 📦 安裝與執行

### 1️⃣ 下載專案

```bash
git clone https://github.com/Sinozakichi/LargeSizeWomanShoes.git
cd LargeSizeWomanShoes
```

### 2️⃣ 設定環境變數

```bash
$env:GO_ENV="debug"
```

請填入資料庫連線資訊等必要設定。

### 3️⃣ 執行爬蟲

```bash
go run main.go daf.go anns.go
```

或使用 Docker：

```bash
docker build -t largeSizeWomanShoes .
```

## 📂 專案目錄結構

```
LARGESIZEWOMANSHOES/
├── .github/ # GitHub 設定與 CI/CD 配置
├── .vscode/ # VS Code launch設定檔
├── css/ # 前端 Template CSS
├── scripts/ # Javascript等靜態資源
├── statics/ # 圖片、HTML等靜態資源
├── .dockerignore # Docker 忽略規則
├── .gitignore # Git 忽略規則
├── anns.go # 爬取 Anns 鞋店的爬蟲邏輯
├── daf.go # 爬取 D+AF 鞋店的爬蟲邏輯
├── Dockerfile # Docker 容器設定檔
├── fly.toml # Fly.io 部署設定檔
├── go.mod # Go 依賴管理
├── go.sum # 依賴版本鎖定檔
├── LICENSE # 授權條款
├── main.go # 主程式入口
└── README.md # 專案說明文件

```

## 🤝 貢獻方式

1. **Fork** 此專案
2. 建立新分支 (`git checkout -b feature/my-feature`)
3. 提交修改 (`git commit -m "新增 XXX 功能"`)
4. 推送到你的 Fork (`git push origin feature/my-feature`)
5. 提交 Pull Request

## 📝 授權條款

本專案依照 **MIT License** 授權，詳見 [LICENSE](LICENSE) 檔案。
