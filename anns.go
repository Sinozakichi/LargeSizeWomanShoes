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
	"strings"
	"sync"
)

type RequestBody struct {
	ShopId        int       `json:"shopId"`
	Lang          string    `json:"lang"`
	OperationName string    `json:"operationName"`
	Query         string    `json:"query"`
	Variables     Variables `json:"variables"`
}

type SizeRequestBody struct {
	Ids                      string `json:"ids"`
	IsShowSaleProductOuterId bool   `json:"isShowSaleProductOuterId"`
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

type AnnsShoeDetailOrignalHTML struct {
	ReturnCode string         `json:"ReturnCode"`
	Data       AnnsShoeDetail `json:"Data"`
	Message    string         `json:"Message"`
}

type AnnsShoeDetail struct {
	Id                   int           `json:"Id"`
	ShopId               int           `json:"ShopId"`
	Title                string        `json:"Title"`
	SaleProductSKUIdList []int         `json:"SaleProductSKUIdList"`
	MajorList            []MajorList   `json:"MajorList"`
	SalePageGroup        SalePageGroup `json:"SalePageGroup"`
}

type MajorList struct {
	Title   string    `json:"Title"`
	Price   float64   `json:"Price"`
	SKUList []SKUList `json:"SKUList"`
}

type SKUList struct {
	Title               string     `json:"Title"`
	PropertyList        []Property `json:"PropertyList"`
	DisplayPropertyName string     `json:"DisplayPropertyName"`
}

type Property struct {
	Name            string `json:"Name"`
	PropertySet     string `json:"PropertySet"`
	PropertyNameSet string `json:"PropertyNameSet"`
}

type SalePageGroup struct {
	GroupCode      string         `json:"GroupCode"`
	GroupTitle     string         `json:"GroupTitle"`
	GroupIconStyle string         `json:"GroupIconStyle"`
	SalePageItems  []SalePageItem `json:"SalePageItems"`
}

type SalePageItem struct {
	SalePageId           int    `json:"SalePageId"`
	GroupItemTitle       string `json:"GroupItemTitle"`
	ItemSort             int    `json:"ItemSort"`
	ItemUrl              string `json:"ItemUrl"`
	ImageUpdatedDateTime string `json:"ImageUpdatedDateTime"`
	ImageGroup           string `json:"ImageGroup"`
}

type SaleProductSKUIdDO struct {
	GoodsSKUId         int    `json:"GoodsSKUId"`
	SellingQty         int    `json:"SellingQty"`
	SaleProductSKUId   int    `json:"SaleProductSKUId"`
	StockQty           int    `json:"StockQty"`
	SaleProductOuterId string `json:"SaleProductOuterId"`
}

const rootAPIURL = "https://fts-api.91app.com/pythia-cdn/graphql"
const childAPIURL = "https://www.anns.tw/webapi/SalePageV2/GetSalePageV2Info/123/"
const salepageURL = "https://www.anns.tw/SalePage/Index/"
const sizeStockAPIURL = "https://www.anns.tw/webapi/ProductStock/GetSellingQtyListNew?v=0&shopId=123&lang=zh-TW"

//const chromePath = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"

func getAnnsFliterResponse(orderby, searchSize, searchColor, searchHeel, searchCat string) ([]Shoe, error) {

	var shoes []Shoe
	var resp *http.Response
	var client *http.Client
	startIndex := 0
	totalSize := 0

	// 記錄參數
	log.Printf("func:getAnnsFliterResponse,Ann's篩選條件 - 排序規則: %s, 尺碼: %s, 顏色: %s, 跟高: %s, 款式: %s", orderby, searchSize, searchColor, searchHeel, searchCat)

	// 將 searchCat 轉換為整數
	categoryId, err := strconv.Atoi(searchCat)
	if err != nil {
		log.Println("func:getAnnsFliterResponse,Ann's CategoryId 轉換錯誤:", err)
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
		log.Println("func:getAnnsFliterResponse,Ann's JSON 編碼錯誤:", err)
		return shoes, err
	}

	if enviroment == "release" {
		// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
		client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
		if err != nil {
			log.Println("func:getAnnsFliterResponse,Ann's 無法創建 HTTP 客戶端:", err)
			return shoes, err
		}

	} else {
		// 本地端，不用設定 CA 憑證
		client = &http.Client{}
	}

	// 帶有 CA 憑證的 HTTP 客戶端向 Ann's 打 Fliter HTTP POST 請求
	resp, err = client.Post(rootAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("func:getAnnsFliterResponse,Ann's 商品列表初始請求錯誤:", err)
		return shoes, err
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("func:getAnnsFliterResponse,Ann's 商品列表初始讀取回應錯誤:", err)
		return shoes, err
	}

	// 提取並解析傳回來body.json的資料
	shoes, totalSize, err = extractSalePageList(body)
	if err != nil {
		log.Println("func:getAnnsFliterResponse,Ann's解析 salePageList 錯誤:", err)
		return shoes, err
	}
	log.Printf("結束請求與解析，從編號%d開始到編號%d", startIndex, startIndex+len(shoes))

	// 拿到totalSize後，再去拿所有鞋子的資訊，因為他一次請求只會回最多100雙，因此要迴圈請求
	startIndex += 100
	if totalSize > startIndex {
		shoes, err = getTotalShoesByFliterResponse(shoes, startIndex, totalSize, requestBody)
	}
	if err != nil {
		log.Println("func:getAnnsFliterResponse,Ann's 去拿所有鞋子的資訊錯誤:", err)
		return shoes, err
	}

	// 遍歷訪問shoes.URL，取得每個shoes的Size和Color
	getSizeAndColorByHttpRequset(shoes)

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
		return nil, 0, fmt.Errorf("func:extractSalePageList,ann's JSON 解析錯誤: %v", err)
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
			log.Printf("開始鞋子List請求，從編號%d開始", startIndex)
			requestBody.Variables.StartIndex = startIndex
			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				log.Println("鞋子List請求,Ann's JSON 編碼錯誤:", err)
				return
			}

			if enviroment == "release" {
				// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
				client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
				if err != nil {
					log.Println("func:getTotalShoesByFliterResponse,Ann's 無法創建 HTTP 客戶端:", err)
					return
				}
			} else {
				// 本地端，不用設定 CA 憑證
				client = &http.Client{}
			}
			// 直接向 Ann's 打 Fliter HTTP POST 請求
			resp, err = client.Post(rootAPIURL, "application/json", bytes.NewBuffer(jsonData))

			if err != nil {
				log.Println("鞋子List請求,Ann's 鞋子List Post請求錯誤:", err)
				return
			}
			defer resp.Body.Close()

			// 讀取回應內容
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("鞋子List請求,Ann's 鞋子List請求讀取Body錯誤:", err)
				return
			}

			// 提取並解析傳回來body.json的資料
			newShoes, totalSize, err = extractSalePageList(body)
			if err != nil {
				log.Println("鞋子List請求，Ann's解析 salePageList 錯誤:", err)
				return
			}
			log.Printf("鞋子List請求，結束請求與解析，從編號%d開始到編號%d", startIndex, startIndex+len(newShoes))

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
func getSizeAndColorByHttpRequset(shoes []Shoe) {

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

	// 用 semaphore 限制同時執行的 goroutine 數量
	var sem = make(chan struct{}, 30) // 限制同時最多 X 個 goroutines

	// 使用 rod 包啟動無頭瀏覽器
	// url := launcher.New().Headless(true).MustLaunch()
	// browser := rod.New().ControlURL(url).MustConnect()
	// log.Println("Ann's Headless瀏覽器已啟動")
	// defer browser.Close() // 確保程式結束時關閉瀏覽器

	for i := range shoes {
		// 增加 WaitGroup 計數
		wg.Add(1)
		go func(i int) {
			// 當 goroutine 完成時減少 WaitGroup 計數
			defer wg.Done()

			var client *http.Client
			var err error

			// 使用 semaphore 保證最大併發數
			sem <- struct{}{}
			defer func() { <-sem }() // 完成後釋放 semaphore

			childURL := childAPIURL + shoes[i].ListID

			if enviroment == "release" {
				// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
				client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
			} else {
				// 本地端，不用設定 CA 憑證
				client = &http.Client{}
			}

			// 發送 GET 請求
			resp, err := client.Get(childURL)
			if err != nil {
				log.Println("取得鞋子尺寸與顏色JSON,Ann's 發Get請求錯誤:", err)
				return
			}
			defer resp.Body.Close()

			// 讀取回應內容
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("取得鞋子尺寸與顏色JSON,Ann's 解析Body錯誤:", err)
				return
			}

			//log.Printf("商品編號:%s, 已成功加載頁面", shoes[i].ListID)

			// 解析 HTML 取得鞋子尺寸與顏色
			size, color, err := extractSizesAndColorsByHttpRequest(body)
			if err != nil {
				log.Printf("取得鞋子尺寸與顏色JSON,Ann's 解析 JSON 異常，商品編號:%s,商品名稱:%s,商品URL:%s，錯誤資訊:%s", shoes[i].ListID, shoes[i].Name, shoes[i].URL, err)
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

// 遍歷訪問shoes.URL，取得每個shoes的Size和Color
func getSizeAndColorByGoRod(shoes []Shoe) {

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

			// 解析 HTML 取得鞋子尺寸與顏色
			size, color, err := extractSizesAndColorsByGoRod(childresp.Body)
			if err != nil {
				log.Printf("Ann's 解析 HTML 異常，商品編號:%s,商品名稱:%s,商品URL:%s，錯誤資訊:%s", shoes[i].ListID, shoes[i].Name, shoes[i].URL, err)
				return
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

// 解析 API 傳回來的資料並從中提取鞋子尺寸跟顏色
func extractSizesAndColorsByHttpRequest(body []byte) ([]string, []string, error) {

	var sizes []string
	var colors []string
	var err error
	var annsShoeDetailOrignalHTML AnnsShoeDetailOrignalHTML
	var annsShoeDetail AnnsShoeDetail

	// 將JSON格式的資料轉型成物件
	err = json.Unmarshal(body, &annsShoeDetailOrignalHTML)
	if err != nil {
		log.Println("Ann's,解析尺寸與顏色的API JSON :", body)
		return sizes, colors, fmt.Errorf("Ann's,解析尺寸與顏色的API,JSON 解析錯誤: %v", err)
	}

	annsShoeDetail = annsShoeDetailOrignalHTML.Data
	log.Printf("Ann's,解析尺寸與顏色的API,商品名:%s", annsShoeDetail.Title)

	// 從 annsShoeDetail 中提取尺寸(下分兩種情況，一種是單色，那他的尺寸是在MajorList[0].SKUList[1]裡，而MajorList[0].SKUList[0]放的是顏色資訊，另一種是多色，那他的尺寸即是在MajorList[0].SKUList[0]裡)
	displayPropertyName := annsShoeDetail.MajorList[0].SKUList[0].DisplayPropertyName
	sizes = strings.Split(displayPropertyName, "/")
	// 檢查 sizes 的長度是否為 1 或者裡面不包含數字
	if len(sizes) == 1 || !containsDigit(sizes) {
		// 檢查 annsShoeDetail.MajorList[0].SKUList[1] 是否存在
		if len(annsShoeDetail.MajorList[0].SKUList) > 1 {
			displayPropertyName = annsShoeDetail.MajorList[0].SKUList[1].DisplayPropertyName
			sizes = strings.Split(displayPropertyName, "/")
		} else {
			log.Printf("Ann's,解析尺寸與顏色的API,商品名:%s非鞋類", annsShoeDetail.Title)
			return sizes, colors, fmt.Errorf("Ann's,解析尺寸與顏色的API,商品非鞋類")
		}
	}

	//篩選出未受罄的尺寸
	sizes = filterStockSizeByHttpRequest(annsShoeDetail.SaleProductSKUIdList, sizes)

	// 從 annsShoeDetail 中提取顏色
	for _, productColor := range annsShoeDetail.SalePageGroup.SalePageItems {
		colors = append(colors, productColor.GroupItemTitle)
	}

	return sizes, colors, nil
}

// 解析 HTML 並從中提取鞋子尺寸跟顏色
func extractSizesAndColorsByGoRod(body io.Reader) ([]string, []string, error) {

	// 讀取 HTML 內容
	htmlBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("ann's 讀取 HTML 失敗: %v", err)
	}
	htmlContent := string(htmlBytes)

	// 找出所有 `PropertyNameSet":"尺寸:XX` 和 `PropertyNameSet":"顏色:XX`
	//sizeRe := regexp.MustCompile(`"PropertyNameSet"\s*:\s*"[^"]*尺寸:(\d+)"`)
	sizeRe := regexp.MustCompile(`"PropertyNameSet"\s*:\s*"[^"]*(?:尺寸|顏色):(\d+)"`)
	sizeMatches := sizeRe.FindAllStringSubmatch(htmlContent, -1)

	if sizeMatches == nil {
		return nil, nil, fmt.Errorf("ann's 未找到任何尺寸資料")
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

	if colorMatches == nil {
		return nil, nil, fmt.Errorf("ann's 未找到任何顏色資料，或其只有單色/沒有顏色")
	}

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

// 篩選出未受罄的尺寸
func filterStockSizeByHttpRequest(saleProductSKUIdList []int, sizes []string) []string {

	var stockSizes []string
	var saleProductSKUIdDO []SaleProductSKUIdDO
	var client *http.Client
	var response *http.Response
	var err error

	// 將 saleProductSKUIdList 轉換為逗號分隔的字符串
	ids := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(saleProductSKUIdList)), ","), "[]")

	sizeRequestBody := SizeRequestBody{
		Ids:                      ids,
		IsShowSaleProductOuterId: false,
	}
	sizeJsonData, err := json.Marshal(sizeRequestBody)
	if err != nil {
		log.Println("篩選出未受罄的尺寸,Ann's JSON 編碼打尺寸資訊的API的請求參數錯誤:", err)
		return nil
	}

	if enviroment == "release" {
		// 正式環境，要設定自訂的帶有 CA 憑證的 HTTP 客戶端
		client, err = createHTTPClientWithCACert("/etc/ssl/certs/ca-certificates.crt")
	} else {
		// 本地端，不用設定 CA 憑證
		client = &http.Client{}
	}

	// 帶有 CA 憑證的 HTTP 客戶端向 Ann's 打 Fliter HTTP POST 請求
	response, err = client.Post(sizeStockAPIURL, "application/json", bytes.NewBuffer(sizeJsonData))
	if err != nil {
		log.Println("篩選出未受罄的尺寸,Ann's 打尺寸資訊的API錯誤:", err)
		return nil
	}
	defer response.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("篩選出未受罄的尺寸,Ann's 讀取打尺寸資訊的API的回應錯誤:", err)
		return nil
	}

	// 解析回應
	err = json.Unmarshal(body, &saleProductSKUIdDO)
	if err != nil {
		log.Println("篩選出未受罄的尺寸,Ann's 尺寸資訊的API的回應的 JSON 解析錯誤:", err)
		return nil
	}

	// 遍歷 saleProductSKUIdDO，挑選出 SellingQty > 0 的元素，並將其對應的尺寸加入 stockSizes
	for i, sku := range saleProductSKUIdDO {
		if sku.SellingQty > 0 {
			stockSizes = append(stockSizes, sizes[i])
		}
	}

	return stockSizes

}
