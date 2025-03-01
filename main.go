package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

type Shoe struct {
	//編號、名稱、圖片、URL、價格、當前有的尺碼、顏色
	ListID string   `json:"listID"`
	Name   string   `json:"name"`
	Image  string   `json:"image"`
	URL    string   `json:"url"`
	Price  string   `json:"price"`
	Size   []string `json:"size"`
	Color  []string `json:"color"`
}

var enviroment string

func main() {

	// Terminal啟動: $env:GO_ENV = "debug"
	// >> go run main.go anns.go daf.go

	// 設定環境變數
	enviroment = os.Getenv("GO_ENV")
	log.Println("GO_ENV:" + enviroment)

	if enviroment == "debug" {
		// 設定靜態文件伺服器(Local)
		log.Println("Debug enviroment")
		staticFs := http.FileServer(http.Dir("./statics"))
		scriptFs := http.FileServer(http.Dir("./scripts"))
		http.Handle("/statics/", http.StripPrefix("/statics/", staticFs))
		http.Handle("/scripts/", http.StripPrefix("/scripts/", scriptFs))

	} else if enviroment == "release" {
		// 設定靜態文件伺服器 (PRD)
		// 處理靜態文件，設置了一個路由來處理以 statics 開頭的請求。http.StripPrefix("/statics/", staticFs) 創建了一個新的處理器，這個處理器會去掉請求 URL 中的 statics 前綴，然後將剩餘部分交給 staticFs 處理。例如，當請求 URL 是 index.html 時，實際上會從 index.html 提供文件。
		log.Println("Release enviroment")
		staticFs := http.FileServer(http.Dir("/app/static"))
		scriptFs := http.FileServer(http.Dir("/app/script"))
		cssFs := http.FileServer(http.Dir("/app/css"))

		http.Handle("/statics/", http.StripPrefix("/statics", staticFs))
		http.Handle("/scripts/", http.StripPrefix("/scripts", scriptFs))
		http.Handle("/css/", http.StripPrefix("/css", cssFs))

	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 預設值
	}

	// 動態生成首頁主頁面
	http.HandleFunc("/", indexHandler)
	// 處理器來處理爬女鞋資訊主請求
	http.HandleFunc("/filter", filterHandler)
	log.Println("伺服器啟動於 http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func filterHandler(w http.ResponseWriter, r *http.Request) {

	//允許跨域請求(CORS)
	w.Header().Set("Access-Control-Allow-Origin", "*") // 允許所有來源
	if r.Method != http.MethodGet {
		http.Error(w, "只接受 GET 請求", http.StatusMethodNotAllowed)
		return
	}

	orderby := r.URL.Query().Get("orderby")
	searchSize := r.URL.Query().Get("searchSize")
	searchColor := r.URL.Query().Get("searchColor")
	searchHeel := r.URL.Query().Get("searchHeel")
	searchCat := r.URL.Query().Get("searchCat")
	store := r.URL.Query().Get("store")

	var shoes []Shoe
	var err error

	log.Println("查詢店鋪:" + store)
	switch store {
	case "daf":
		shoes, err = getDAFFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat)
	case "anns":
		shoes, err = getAnnsFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat)
	default:
		http.Error(w, "未知的商店", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 返回 JSON 結果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shoes)

}

// indexHandler 動態生成 HTML 頁面
func indexHandler(w http.ResponseWriter, r *http.Request) {

	var tmpl *template.Template
	if enviroment == "release" {
		tmpl = template.Must(template.ParseFiles("/app/static/index.html"))
	} else {
		tmpl = template.Must(template.ParseFiles("index.html"))
	}
	data := struct {
		Environment string
	}{
		Environment: enviroment,
	}
	tmpl.Execute(w, data)

	if enviroment == "release" {
		http.Redirect(w, r, "/statics/index.html", http.StatusFound)
	}
}

// createHTTPClientWithCACert 創建一個帶有 CA 憑證的 HTTP 客戶端
func createHTTPClientWithCACert(caCertPath string) (*http.Client, error) {
	// 讀取系統 CA 憑證
	caCertPool := x509.NewCertPool()
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("無法讀取 CA 憑證: %v", err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	// 設定自訂的 http.Client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: caCertPool},
		},
	}
	return client, nil
}

// 檢查切片中是否包含數字的輔助函數
func containsDigit(sizes []string) bool {
	for _, size := range sizes {
		if _, err := strconv.Atoi(size); err == nil {
			return true
		}
	}
	return false
}
