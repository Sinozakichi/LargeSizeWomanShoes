package main

import (
	"fmt"
	"log"
	"net/http"
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
	http.HandleFunc("/filter", filterHandler)
	log.Println("伺服器啟動於 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

	switch store {
	case "daf":
		getDAFFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat)
	default:
	}

	fmt.Fprintf(w, "篩選條件已處理")
}
