package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

const rootURL = "https://www.daf-shoes.com/"

// 靴類的searchCat
var bootCategory = map[string]int{
	"148": 1,
	"199": 2,
	"314": 3,
	"256": 4,
	"259": 5,
}

func getDAFFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat string) ([]Shoe, error) {

	var url string
	var pagecount int = 1
	var isBoot bool = false
	var resp *http.Response
	var err error
	// 用於等待所有 goroutines 完成
	var wg sync.WaitGroup
	// 用於保護共享資源
	var mu sync.Mutex
	shoes := []Shoe{}

	// 記錄參數
	log.Printf("D+AF篩選條件 - 排序規則: %s, 尺碼: %s, 顏色: %s, 跟高: %s, 款式: %s", orderby, searchSize, searchColor, searchHeel, searchCat)

	var fliterQuery = fmt.Sprintf("orderby=%s&searchSize=%s&searchColor=%s&searchHeel=%s&searchCat=%s", orderby, searchSize, searchColor, searchHeel, searchCat)

	// 靴類要打另一個URL
	if _, exists := bootCategory[searchCat]; exists {
		url = fmt.Sprintf("%sproduct/list/303/%d?%s", rootURL, pagecount, fliterQuery)
		isBoot = true
	} else {
		// 其他品項則手動組裝篩選 URL
		url = fmt.Sprintf("%sproduct/list/all/%d?%s", rootURL, pagecount, fliterQuery)
	}
	log.Println("url:" + url)

	if enviroment == "release" {
		// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
		client, err := createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
		if err != nil {
			log.Println("D+AF 無法創建 HTTP 客戶端:", err)
			return shoes, err
		}

		// 帶有 CA 憑證的 HTTP 客戶端向 D+AF 打 Fliter HTTP GET 請求
		resp, err = client.Get(url)
	} else {
		// 本地端，不用設定 CA 憑證
		// 直接向 D+AF 打 Fliter HTTP GET 請求
		resp, err = http.Get(url)
	}

	if err != nil {
		log.Println("D+AF 商品列表初始請求錯誤:", err)
		return shoes, err
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("D+AF 商品列表初始讀取回應錯誤:", err)
		return shoes, err
	}

	// 取出totalPage
	totalPage, err := getTotalPage(body)
	if err != nil {
		log.Println("D+AF 取得totalPage錯誤:", err)
		return shoes, err
	}

	log.Printf("已拿到totalpage，要取全部篩選的鞋子，D+AF 總頁數: %d", totalPage)
	// 拿到totalPage後，遍歷每一頁的商品將之加入shoes
	err = getTotalShoes(totalPage, rootURL, pagecount, fliterQuery, isBoot, &shoes)
	if err != nil {
		log.Println("D+AF 取得所有鞋子錯誤:", err)
		return shoes, err
	}
	log.Printf("已拿取全部篩選的鞋子，總鞋子數: %d", len(shoes))

	// 傳遞結果的 channel
	ch := make(chan struct {
		index int
		size  []string
		color []string
	})

	// 遍歷訪問shoes.URL，取得每個頁面的內容
	for i := range shoes {
		// 增加 WaitGroup 計數
		wg.Add(1)
		go func(i int) {
			// 當 goroutine 完成時減少 WaitGroup 計數
			defer wg.Done()
			// 發送shoes.URL HTTP GET 請求
			childresp, err := http.Get(shoes[i].URL)
			if err != nil {
				log.Println("D+AF 遍歷訪問各商品時請求錯誤:", err)
				return
			}
			defer childresp.Body.Close()

			// 讀取回應內容
			childbody, err := io.ReadAll(childresp.Body)
			if err != nil {
				log.Println("D+AF 遍歷訪問各商品時讀取回應錯誤:", err)
				return
			}

			// 尺碼
			getSize(childbody, &shoes[i])
			// 顏色
			getColor(childbody, &shoes[i])

			// 將結果發送到 channel
			ch <- struct {
				index int
				size  []string
				color []string
			}{index: i, size: shoes[i].Size, color: shoes[i].Color}
		}(i)
	}

	// 啟動一個 goroutine 來等待所有工作完成並關閉 channel
	go func() {
		// 等待所有 goroutines 完成
		wg.Wait()
		close(ch)
	}()

	// 從 channel 接收結果並更新鞋子的尺寸和顏色
	for result := range ch {
		// 鎖定 mutex 以保護共享資源
		mu.Lock()
		shoes[result.index].Size = result.size
		shoes[result.index].Color = result.color
		mu.Unlock()
	}

	return shoes, nil
}

// 從吐回來的Body中取出totalPage
func getTotalPage(body []byte) (int, error) {

	var totalpage int
	var err error

	// 使用正則表達式提取 <input type="hidden" name="totalpage"> 元素的 value 屬性
	re := regexp.MustCompile(`<input[^>]+type=['"]hidden['"][^>]+name=['"]totalpage['"][^>]+value=['"](\d+)['"][^>]*>`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		log.Println("D+AF 未找到匹配的 totalpage")
		return 0, errors.New("D+AF 未找到匹配的 totalpage")
	}

	// 轉型成int
	totalpage, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, errors.New("D+AF totalpage轉換錯誤")
	}
	return totalpage, nil
}

// 依totalPage去取出所有鞋
func getTotalShoes(totalPage int, rootURL string, pagecount int, fliterQuery string, isBoot bool, shoes *[]Shoe) error {

	// 依totalPage去取出所有鞋
	for count := pagecount; count <= totalPage; count++ {

		var resp *http.Response
		var err error
		var url string

		if isBoot {
			url = fmt.Sprintf("%sproduct/list/303/%d?%s", rootURL, count, fliterQuery)
		} else {
			url = fmt.Sprintf("%sproduct/list/all/%d?%s", rootURL, count, fliterQuery)
		}

		log.Println("url:" + url)

		if enviroment == "release" {
			// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
			client, err := createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
			if err != nil {
				log.Println("D+AF totalPage去取出所有鞋無法創建 HTTP 客戶端:", err)
				return err
			}

			// 帶有 CA 憑證的 HTTP 客戶端向 D+AF 打 Fliter HTTP GET 請求
			resp, err = client.Get(url)
		} else {
			// 本地端，不用設定 CA 憑證
			// 直接向 D+AF 打 Fliter HTTP GET 請求
			resp, err = http.Get(url)
		}

		// 向 D+AF 打 Fliter HTTP GET 請求
		resp, err = http.Get(url)
		if err != nil {
			log.Println("D+AF totalPage去取出所有鞋請求錯誤:", err)
			return err
		}
		defer resp.Body.Close()

		// 讀取回應內容
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("D+AF totalPage去取出所有鞋讀取Body錯誤:", err)
			return err
		}

		// ListID、名稱、價格
		getListIDAndNameAndPrize(body, shoes)
		// URL
		getURL(body, shoes, len(*shoes))
		// 圖檔
		getImage(body, shoes, len(*shoes))
	}

	return nil
}

// 從吐回來的Body中取出所有鞋的名稱、價格、數量
func getListIDAndNameAndPrize(body []byte, shoes *[]Shoe) []Shoe {

	// 使用正則表達式提取 JavaScript 物件
	re := regexp.MustCompile(`gtag\('event', 'view_item_list', {[\s\S]+?}\);`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		log.Println("D+AF 未找到匹配的 JavaScript 物件")
		return *shoes
	}

	// 提取 items 部分
	reItems := regexp.MustCompile(`"items": \[([^\]]+)\]`)
	itemsMatch := reItems.FindStringSubmatch(matches[0])
	if len(itemsMatch) == 0 {
		log.Println("D+AF 未找到 items 部分")
		return *shoes
	}

	// 解析 items 部分
	var items []map[string]interface{}
	_ = json.Unmarshal([]byte(fmt.Sprintf("[%s]", itemsMatch[1])), &items)

	// 將 items 轉換為 Shoe 結構體
	for _, item := range items {
		shoe := Shoe{
			ListID: fmt.Sprintf("%v", item["id"]),
			Name:   item["name"].(string),
			Price:  fmt.Sprintf("%v", item["price"]),
		}
		*shoes = append(*shoes, shoe)
	}
	return *shoes
}

// 從吐回來的Body中取出所有鞋的圖檔
func getImage(body []byte, shoes *[]Shoe, num int) []Shoe {

	// 使用正則表達式提取 <source> 標籤中的 srcset 和 id 屬性值
	re := regexp.MustCompile(`<source[^>]+srcset="([^"]+)"[^>]+id="pic(\d+_\d+)w"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	// 將 srcset 值存儲到 Shoe 結構體的 Image 字段
	for _, match := range matches {
		for i := range *shoes {
			if (*shoes)[i].ListID == match[2] {
				(*shoes)[i].Image = match[1]
			}
		}
	}

	return *shoes
}

// 從吐回來的Body中取出所有鞋的URL
func getURL(body []byte, shoes *[]Shoe, num int) []Shoe {

	// 使用正則表達式提取 <a> 標籤中的 href 屬性值
	re := regexp.MustCompile(`<a[^>]*alt="[^"]*"[^>]*href="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	// 將 href 值存儲到 Shoe 結構體的 URL 字段
	for _, match := range matches {
		// 使用正則表達式提取 href 中的後面兩段
		hrefRe := regexp.MustCompile(`/product/show/(\d+)/(\d+)/`)
		hrefMatches := hrefRe.FindStringSubmatch(match[1])
		listID := fmt.Sprintf("%s_%s", hrefMatches[1], hrefMatches[2])
		for i := range *shoes {
			if (*shoes)[i].ListID == listID {
				(*shoes)[i].URL = rootURL + match[1]
			}
		}
	}

	return *shoes
}

// 遍歷每個產品後，從吐回來的Body中取出一雙鞋的尺碼List
func getSize(body []byte, shoe *Shoe) *Shoe {

	// 使用正則表達式提取 <div class='mini-box sizeSel' btn='ok'> 標籤中的文本
	re := regexp.MustCompile(`<div[^>]+class=['"][^'"]*mini-box\s+sizeSel[^'"]*['"][^>]+btn=['"]ok['"][^>]*>.*?</div>`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		log.Printf("D+AF 未找到匹配的 <div class='mini-box sizeSel'> 標籤，應為售罄，商品名稱: %s", shoe.Name)
		return shoe
	}

	// 使用正則表達式提取 <span> 標籤中的文本
	spanRe := regexp.MustCompile(`<span>([^<]+)</span>`)

	// 將匹配到的文本存儲到 Shoe 結構體的 Size 字段
	for _, match := range matches {
		spanMatches := spanRe.FindStringSubmatch(match[0])
		if len(spanMatches) > 1 {
			shoe.Size = append(shoe.Size, spanMatches[1])
		}
	}

	return shoe
}

// 遍歷每個產品後，從吐回來的Body中取出一雙鞋的顏色List
func getColor(body []byte, shoe *Shoe) *Shoe {

	// 使用正則表達式提取 <div class='mini-box color colorSel' title="顏色名稱"> 標籤中的顏色名稱
	re := regexp.MustCompile(`<div[^>]+class=['"][^'"]*mini-box\s+color\s+colorSel[^'"]*['"][^>]+title=['"]([^'"]+)['"][^>]*>`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		log.Printf("D+AF 未找到匹配的 <div class='mini-box colorSel'> 標籤，商品名稱: %s", shoe.Name)
		return shoe
	}

	// 將顏色名稱添加到 Shoe 結構體的 Color 字段
	for _, match := range matches {
		if len(match) > 1 {
			shoe.Color = append(shoe.Color, match[1]) // match[1] 是 title 屬性中的顏色名稱
		}
	}

	return shoe
}
