package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

type RequestBody struct {
	ShopId        int       `json:"shopId"`
	Lang          string    `json:"lang"`
	OperationName string    `json:"operationName"`
	Query         string    `json:"query"`
	Variables     Variables `json:"variables"`
}

type Variables struct {
	ShopId               int         `json:"shopId"`
	CategoryId           int         `json:"categoryId"`
	StartIndex           int         `json:"startIndex"`
	FetchCount           int         `json:"fetchCount"`
	OrderBy              string      `json:"orderBy"`
	IsShowCurator        bool        `json:"isShowCurator"`
	TagFilters           []TagFilter `json:"tagFilters,omitempty"`
	TagShowMore          bool        `json:"tagShowMore"`
	MinPrice             interface{} `json:"minPrice"`
	MaxPrice             interface{} `json:"maxPrice"`
	PayType              []string    `json:"payType"`
	ShippingType         []string    `json:"shippingType"`
	IncludeSalePageGroup bool        `json:"includeSalePageGroup"`
	LocationId           interface{} `json:"locationId"`
}

type TagFilter struct {
	GroupId string `json:"groupId"`
	KeyId   string `json:"keyId"`
}

type ResponseData struct {
	Data struct {
		ShopCategory struct {
			SalePageList struct {
				SalePageList     []AnnsShoe `json:"salePageList"`
				TotalSize        int        `json:"totalSize"`
				ShopCategoryId   int        `json:"shopCategoryId"`
				ShopCategoryName string     `json:"shopCategoryName"`
			} `json:"salePageList"`
		} `json:"shopCategory"`
	} `json:"data"`
}

type AnnsShoe struct {
	SalePageId int      `json:"salePageId"`
	Title      string   `json:"title"`
	PicUrl     string   `json:"picUrl"`
	PicList    []string `json:"picList"`
	Price      int      `json:"price"`
}

type SKUProperty struct {
	GoodsSKUId       int     `json:"GoodsSKUId"`
	PropertySet      string  `json:"PropertySet"`
	SaleProductSKUId int     `json:"SaleProductSKUId"`
	SellingQty       int     `json:"SellingQty"`
	OnceQty          int     `json:"OnceQty"`
	PropertyNameSet  string  `json:"PropertyNameSet"`
	IsShow           bool    `json:"IsShow"`
	Price            float64 `json:"Price"`
	SuggestPrice     float64 `json:"SuggestPrice"`
	CartonQty        int     `json:"CartonQty"`
}

type SalePageIndexViewModel struct {
	Id                 int           `json:"Id"`
	ShopId             int           `json:"ShopId"`
	ShopName           string        `json:"ShopName"`
	ShopCategoryId     int           `json:"ShopCategoryId"`
	CategoryId         int           `json:"CategoryId"`
	CategoryName       string        `json:"CategoryName"`
	SKUPropertySetList []SKUProperty `json:"SKUPropertySetList"`
}

const rootAPIURL = "https://fts-api.91app.com/pythia-cdn/graphql"
const salepageURL = "https://www.anns.tw/SalePage/Index/"

func getAnnsFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat string) ([]Shoe, error) {

	var shoes []Shoe
	var resp *http.Response
	var client *http.Client
	startIndex := 0
	totalSize := 0

	// 記錄參數
	log.Printf("Ann's篩選條件 - 排序規則: %s, 尺碼: %s, 顏色: %s, 跟高: %s, 款式: %s", orderby, searchSize, searchColor, searchHeel, searchCat)

	// 將 searchCat 轉換為整數
	categoryId, err := strconv.Atoi(searchCat)
	if err != nil {
		log.Println("Ann's CategoryId 轉換錯誤:", err)
		return shoes, err
	}
	// 構建請求的 Body
	tagFilters := []TagFilter{}
	if searchColor != "" {
		tagFilters = append(tagFilters, TagFilter{GroupId: "G87", KeyId: searchColor})
	}
	if searchHeel != "" {
		tagFilters = append(tagFilters, TagFilter{GroupId: "G88", KeyId: searchHeel})
	}

	requestBody := RequestBody{
		ShopId:        123,
		Lang:          "zh-TW",
		OperationName: "cms_shopCategory",
		Query:         "query cms_shopCategory($shopId: Int!, $categoryId: Int!, $startIndex: Int!, $fetchCount: Int!, $orderBy: String, $isShowCurator: Boolean, $locationId: Int, $tagFilters: [ItemTagFilter], $tagShowMore: Boolean, $serviceType: String, $minPrice: Float, $maxPrice: Float, $payType: [String], $shippingType: [String], $includeSalePageGroup: Boolean) {\n  shopCategory(shopId: $shopId, categoryId: $categoryId) {\n    salePageList(startIndex: $startIndex, maxCount: $fetchCount, orderBy: $orderBy, isCuratorable: $isShowCurator, locationId: $locationId, tagFilters: $tagFilters, tagShowMore: $tagShowMore, minPrice: $minPrice, maxPrice: $maxPrice, payType: $payType, shippingType: $shippingType, serviceType: $serviceType, includeSalePageGroup: $includeSalePageGroup) {\n      salePageList {\n        salePageId\n        title\n        picUrl\n        picList\n        salePageCode\n        price\n        suggestPrice\n        isFav\n        isComingSoon\n        isSoldOut\n        soldOutActionType\n        sellingQty\n        pairsPoints\n        pairsPrice\n        priceDisplayType\n        displayTags {\n          group\n          keys {\n            id\n            startTime\n            endTime\n            picUrl {\n              ratioOneToOne\n              ratioThreeToFour\n              __typename\n            }\n            __typename\n          }\n          __typename\n        }\n        salePageGroup {\n          groupTitle\n          groupIconStyle\n          groupItems {\n            salePageId\n            itemTitle\n            itemUrl\n            __typename\n          }\n          __typename\n        }\n        promotionPrices {\n          promotionEngineId\n          memberCollectionId\n          price\n          startDateTime\n          endDateTime\n          label\n          __typename\n        }\n        isRestricted\n        enableIsComingSoon\n        isShowSellingStartDateTime\n        sellingStartDateTime\n        listingStartDateTime\n        metafields\n        __typename\n      }\n      totalSize\n      shopCategoryId\n      shopCategoryName\n      statusDef\n      listModeDef\n      orderByDef\n      dataSource\n      tags {\n        isGroupShowMore\n        groups {\n          groupId\n          groupDisplayName\n          isKeyShowMore\n          keys {\n            keyId\n            keyDisplayName\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      priceRange {\n        min\n        max\n        __typename\n      }\n      __typename\n    }\n  }\n}",
		Variables: Variables{
			ShopId:               123,
			CategoryId:           categoryId,
			StartIndex:           startIndex,
			FetchCount:           600,
			OrderBy:              orderby,
			IsShowCurator:        true,
			TagFilters:           tagFilters,
			TagShowMore:          true,
			MinPrice:             nil,
			MaxPrice:             nil,
			PayType:              []string{},
			ShippingType:         []string{},
			IncludeSalePageGroup: true,
			LocationId:           nil,
		},
	}

	log.Printf("開始請求，從編號%d開始", startIndex)
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Ann's JSON 編碼錯誤:", err)
		return shoes, err
	}

	if enviroment == "release" {
		// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
		client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
		if err != nil {
			log.Println("Ann's 無法創建 HTTP 客戶端:", err)
			return shoes, err
		}

	} else {
		// 本地端，不用設定 CA 憑證
		client = &http.Client{}
	}

	// 帶有 CA 憑證的 HTTP 客戶端向 Ann's 打 Fliter HTTP POST 請求
	resp, err = client.Post(rootAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Ann's 商品列表初始請求錯誤:", err)
		return shoes, err
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ann's 商品列表初始讀取回應錯誤:", err)
		return shoes, err
	}

	// 提取並解析傳回來body.json的資料
	shoes, totalSize, err = extractSalePageList(body)
	if err != nil {
		log.Println("Ann's解析 salePageList 錯誤:", err)
		return shoes, err
	}
	log.Printf("結束請求與解析，從編號%d開始到編號%d", startIndex, startIndex+len(shoes))

	// 拿到totalSize後，再去拿所有鞋子的資訊，因為他一次請求只會回最多100雙，因此要迴圈請求
	startIndex += 100
	if totalSize > startIndex {
		shoes, err = getTotalShoesByFliterResponse(shoes, startIndex, totalSize, requestBody)
	}
	if err != nil {
		log.Println("Ann's 去拿所有鞋子的資訊錯誤:", err)
		return shoes, err
	}

	// 遍歷訪問shoes.URL，取得每個shoes的Size和Color
	getSizeAndColor(shoes)

	// 篩選出有符合尺寸的鞋子
	filteredShoes := filterShoesBySize(shoes, searchSize)
	log.Printf("結束尺寸篩選，共有%d雙鞋", len(filteredShoes))

	return filteredShoes, nil
}

// 提取並解析傳回來body.json的資料，並塞入ListID、Name、Price、Image、URL
func extractSalePageList(body []byte) ([]Shoe, int, error) {
	var responseData ResponseData
	err := json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, 0, fmt.Errorf("Ann's JSON 解析錯誤: %v", err)
	}

	var shoes []Shoe
	for _, item := range responseData.Data.ShopCategory.SalePageList.SalePageList {
		// 將 SalePageId 和 Price 轉換為string
		salePageId := fmt.Sprintf("%v", item.SalePageId)
		price := fmt.Sprintf("%v", item.Price)

		shoe := Shoe{
			ListID: salePageId,
			Name:   item.Title,
			Image:  item.PicUrl,
			URL:    salepageURL + salePageId,
			Price:  price,
		}
		shoes = append(shoes, shoe)
	}
	totalSize := responseData.Data.ShopCategory.SalePageList.TotalSize

	return shoes, totalSize, nil
}

// 拿到totalSize後，再去拿所有鞋子的資訊，因為他一次請求只會回最多100雙
// 注意:在併發區塊下下斷點，可能會有系統錯誤!
func getTotalShoesByFliterResponse(shoes []Shoe, startIndex, totalSize int, requestBody RequestBody) ([]Shoe, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	ch := make(chan []Shoe)

	// 發送並發請求
	// 每次請求100筆資料
	for ; startIndex < totalSize; startIndex += 100 {
		wg.Add(1)
		go func(startIndex int) {

			var resp *http.Response
			var client *http.Client
			var newShoes []Shoe
			defer wg.Done()

			// 更新 requestBody 中的 StartIndex
			log.Printf("開始請求，從編號%d開始", startIndex)
			requestBody.Variables.StartIndex = startIndex
			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				log.Println("Ann's JSON 編碼錯誤:", err)
				return
			}

			if enviroment == "release" {
				// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
				client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
				if err != nil {
					log.Println("Ann's 無法創建 HTTP 客戶端:", err)
					return
				}
			} else {
				// 本地端，不用設定 CA 憑證
				client = &http.Client{}
			}
			// 直接向 Ann's 打 Fliter HTTP POST 請求
			resp, err = client.Post(rootAPIURL, "application/json", bytes.NewBuffer(jsonData))

			if err != nil {
				log.Println("Ann's 商品列表初始請求錯誤:", err)
				return
			}
			defer resp.Body.Close()

			// 讀取回應內容
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Ann's 商品列表初始讀取回應錯誤:", err)
				return
			}

			// 提取並解析傳回來body.json的資料
			newShoes, totalSize, err = extractSalePageList(body)
			if err != nil {
				log.Println("Ann's解析 salePageList 錯誤:", err)
				return
			}
			log.Printf("結束請求與解析，從編號%d開始到編號%d", startIndex, startIndex+len(newShoes))

			ch <- newShoes

		}(startIndex)
	}

	// 啟動一個 goroutine 來等待所有工作完成並關閉 channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// 從 channel 接收結果並更新鞋子列表
	for newShoes := range ch {
		mu.Lock()
		shoes = append(shoes, newShoes...)
		mu.Unlock()
	}

	log.Println("撈取鞋子總雙數:", len(shoes))
	return shoes, nil
}

// 遍歷訪問shoes.URL，取得每個shoes的Size和Color
func getSizeAndColor(shoes []Shoe) {

	// 用於等待所有 goroutines 完成
	var wg sync.WaitGroup //類似C#的Task
	// 用於保護共享資源
	var mu sync.Mutex
	// 傳遞結果的 channel
	ch := make(chan struct {
		index int
		size  []string
		color []string
	})

	log.Println("要訪問的鞋子總雙數:", len(shoes))
	for i := range shoes {
		// 增加 WaitGroup 計數
		wg.Add(1)
		go func(i int) {
			// 當 goroutine 完成時減少 WaitGroup 計數
			defer wg.Done()

			var client *http.Client
			var err error

			if enviroment == "release" {
				// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
				client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
				if err != nil {
					log.Println("Ann's 無法創建 HTTP 客戶端:", err)
					return
				}
			} else {
				// 本地端，不用設定 CA 憑證
				client = &http.Client{}
			}

			// 發送shoes.URL HTTP GET 請求
			childresp, err := client.Get(shoes[i].URL)
			if err != nil {
				log.Println("Ann's 遍歷訪問各商品時請求錯誤:", err)
				return
			}
			defer childresp.Body.Close()

			// 設定標頭模擬正常瀏覽器
			// req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
			// req.Header.Set("Referer", "https://www.anns.tw/") // 可視情況調整 Referer

			// // 送出請求
			// childresp, err := client.Do(req)
			// if err != nil {
			// 	log.Println("Ann's 遍歷訪問各商品時請求錯誤:", err)
			// 	return
			// }
			// defer childresp.Body.Close()

			// 讀取 HTML 內容
			htmlBytes, err := io.ReadAll(childresp.Body)
			if err != nil {
				log.Println("Ann's 讀取 HTML 失敗:", err)
				return
			}
			htmlContent := string(htmlBytes)
			log.Printf("商品編號:%s,HTTP StatusCode:%d", shoes[i].ListID, childresp.StatusCode)

			// 解析 HTML 取得鞋子尺寸與顏色
			size, color, err := extractSizesAndColors(htmlContent)
			if err != nil {
				log.Printf("Ann's 解析 HTML 異常，商品編號:%s,商品名稱:%s,商品URL:%s，錯誤資訊:%s", shoes[i].ListID, shoes[i].Name, shoes[i].URL, err)
				// log.Printf("寫入 debug.html紀錄，debug_%s.html", shoes[i].ListID)

				// // 設定檔案名稱
				// debugFilename := fmt.Sprintf("debug_%s.html", shoes[i].ListID)

				// // 寫入 debug_{ListID}.html
				// writeErr := os.WriteFile(debugFilename, htmlBytes, 0644)
				// if writeErr != nil {
				// 	log.Printf("無法寫入 %s: %v", debugFilename, writeErr)
				// }
				// return
			}

			// 將結果發送到 channel
			ch <- struct {
				index int
				size  []string
				color []string
			}{index: i, size: size, color: color}
		}(i)
	}

	// 啟動一個 goroutine 來等待所有工作完成並關閉 channel
	go func() {
		// 等待所有 goroutines 完成
		wg.Wait() //類似C#的Task.WaitAll()
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
}

// 解析 HTML 並從中提取鞋子尺寸跟顏色
func extractSizesAndColors(htmlContent string) ([]string, []string, error) {

	// 找出所有 `PropertyNameSet":"尺寸:XX` 和 `PropertyNameSet":"顏色:XX`
	//sizeRe := regexp.MustCompile(`"PropertyNameSet"\s*:\s*"[^"]*尺寸:(\d+)"`)
	sizeRe := regexp.MustCompile(`"PropertyNameSet"\s*:\s*"[^"]*(?:尺寸|顏色):(\d+)"`)
	sizeMatches := sizeRe.FindAllStringSubmatch(htmlContent, -1)

	if sizeMatches == nil {
		return nil, nil, fmt.Errorf("Ann's 未找到任何尺寸資料")
	}

	// 用 map 避免重複
	sizeSet := make(map[string]struct{})
	for _, match := range sizeMatches {
		sizeSet[match[1]] = struct{}{}
	}

	// 轉成 slice 回傳
	sizes := make([]string, 0, len(sizeSet))
	for size := range sizeSet {
		sizes = append(sizes, size)
	}
	// 找出所有 "GroupItemTitle": 後的顏色
	colorRe := regexp.MustCompile(`"GroupItemTitle"\s*:\s*"([^"]*)"`)
	colorMatches := colorRe.FindAllStringSubmatch(htmlContent, -1)

	// if colorMatches == nil {
	// 	return nil, nil, fmt.Errorf("Ann's 未找到任何顏色資料，或其只有單色/沒有顏色")
	// }

	// 用 map 避免重複
	colorSet := make(map[string]struct{})
	for _, match := range colorMatches {
		colorSet[match[1]] = struct{}{}
	}

	// 轉成 slice 回傳
	colors := make([]string, 0, len(colorSet))
	for color := range colorSet {
		colors = append(colors, color)
	}
	return sizes, colors, nil
}

// 尺寸篩選
func filterShoesBySize(shoes []Shoe, searchSize string) []Shoe {
	var filteredShoes []Shoe
	for _, shoe := range shoes {
		for _, size := range shoe.Size {
			if size == searchSize {
				filteredShoes = append(filteredShoes, shoe)
			}
		}
	}
	return filteredShoes
}
