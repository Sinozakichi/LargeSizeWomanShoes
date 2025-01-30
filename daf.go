package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

const rootURL = "https://www.daf-shoes.com/"

func getDAFFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat string) {

	// 記錄參數
	log.Printf("D+AF篩選條件 - 排序規則: %s, 尺碼: %s, 顏色: %s, 跟高: %s, 款式: %s", orderby, searchSize, searchColor, searchHeel, searchCat)

	// 手動組裝篩選 URL
	url := fmt.Sprintf("%sproduct/list/all?orderby=%s&searchSize=%s&searchColor=%s&searchHeel=%s&searchCat=%s", rootURL, orderby, searchSize, searchColor, searchHeel, searchCat)

	// 向 D+AF 打 Fliter HTTP GET 請求
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("請求錯誤:", err)
		return
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("讀取回應錯誤:", err)
		return
	}

	shoes := []Shoe{}
	// 取出訊息
	// ListID、名稱、價格
	getListIDAndNameAndPrize(body, &shoes)
	// URL
	getURL(body, &shoes, len(shoes))
	// 圖檔
	getImage(body, &shoes, len(shoes))

	// 遍歷訪問shoes.URL，取得每個頁面的內容
	for i := range shoes {
		childresp, err := http.Get(shoes[i].URL)
		if err != nil {
			fmt.Println("請求錯誤:", err)
			return
		}
		defer childresp.Body.Close()

		// 讀取回應內容
		childbody, err := io.ReadAll(childresp.Body)
		if err != nil {
			fmt.Println("讀取回應錯誤:", err)
			return
		}

		// 尺碼
		getSize(childbody, &shoes[i])
		// 顏色
		getColor(childbody, &shoes[i])
	}

	// 驗證篩選出的資料
	for _, shoe := range shoes {
		fmt.Printf("ListID:%s, Name: %s, Price: %s,Image: %s,URL:%s", shoe.ListID, shoe.Name, shoe.Price, shoe.Image, shoe.URL)
	}
}

// 從吐回來的Body中取出所有鞋的名稱、價格、數量
func getListIDAndNameAndPrize(body []byte, shoes *[]Shoe) []Shoe {

	// 使用正則表達式提取 JavaScript 物件
	re := regexp.MustCompile(`gtag\('event', 'view_item_list', {[\s\S]+?}\);`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		fmt.Println("未找到匹配的 JavaScript 物件")
		return *shoes
	}

	// 提取 items 部分
	reItems := regexp.MustCompile(`"items": \[([^\]]+)\]`)
	itemsMatch := reItems.FindStringSubmatch(matches[0])
	if len(itemsMatch) == 0 {
		fmt.Println("未找到 items 部分")
		return *shoes
	}

	// 解析 items 部分
	var items []map[string]interface{}
	_ = json.Unmarshal([]byte(fmt.Sprintf("[%s]", itemsMatch[1])), &items)

	// 將 items 轉換為 Shoe 結構體
	for _, item := range items {
		shoe := Shoe{
			ListID: fmt.Sprintf("%v", item["list_position"]),
			Name:   item["name"].(string),
			Price:  fmt.Sprintf("%v", item["price"]),
		}
		*shoes = append(*shoes, shoe)
	}
	return *shoes
}

// 從吐回來的Body中取出所有鞋的圖檔
func getImage(body []byte, shoes *[]Shoe, num int) []Shoe {

	// 使用正則表達式提取 <source> 標籤中的 srcset 屬性值
	re := regexp.MustCompile(`<source[^>]+srcset="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(body), num)
	if len(matches) != num {
		fmt.Println("未完全匹配正確的 <source> 標籤")
		return *shoes
	}
	// 將 srcset 值存儲到 Shoe 結構體的 Image 字段
	for i, match := range matches {
		if i < num {
			(*shoes)[i].Image = match[1]
		}
	}

	return *shoes
}

// 從吐回來的Body中取出所有鞋的URL
func getURL(body []byte, shoes *[]Shoe, num int) []Shoe {

	// 使用正則表達式提取 <a> 標籤中的 href 屬性值
	re := regexp.MustCompile(`<a[^>]*alt="[^"]*"[^>]*href="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(body), num)
	if len(matches) != num {
		fmt.Println("未完全匹配正確的 <a> 標籤")
		return *shoes
	}

	// 將 href 值存儲到 Shoe 結構體的 URL 字段
	for i, match := range matches {
		if i < num {
			(*shoes)[i].URL = rootURL + match[1]
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
		fmt.Println("未找到匹配的 <div class='mini-box sizeSel'> 標籤")
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
		fmt.Println("未找到匹配的 <div class='mini-box colorSel'> 標籤")
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
