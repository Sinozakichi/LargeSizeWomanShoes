package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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

func main() {

	// 設定靜態文件伺服器
	// staticFs := http.FileServer(http.Dir("./static"))
	// scriptFs := http.FileServer(http.Dir("./script"))
	// http.Handle("/static/", http.StripPrefix("/static/", staticFs))
	// http.Handle("/script/", http.StripPrefix("/script/", scriptFs))

	// 設定靜態文件伺服器
	staticFs := http.FileServer(http.Dir("/app/static"))
	scriptFs := http.FileServer(http.Dir("/app/script"))

	// 處理靜態文件
	http.Handle("/static/", http.StripPrefix("/static", staticFs))
	http.Handle("/script/", http.StripPrefix("/script", scriptFs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 預設值
	}

	http.HandleFunc("/filter", filterHandler)
	log.Println("伺服器啟動於 http://localhost:8080")
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
