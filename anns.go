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
				shopCategoryName string     `json:"shopCategoryName"`
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

	// 記錄參數
	log.Printf("Ann's篩選條件 - 排序規則: %s, 尺碼: %s, 顏色: %s, 跟高: %s, 款式: %s", orderby, searchSize, searchColor, searchHeel, searchCat)

	// 要先打一支API去拿所有鞋的List

	// 將 searchCat 轉換為整數
	categoryId, err := strconv.Atoi(searchCat)
	if err != nil {
		fmt.Println("CategoryId 轉換錯誤:", err)
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
			StartIndex:           0,
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

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("JSON 編碼錯誤:", err)
		return shoes, err
	}

	// 向 Ann's 打 Fliter HTTP POST 請求
	resp, err := http.Post(rootAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("初始請求錯誤:", err)
		return shoes, err
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("初始讀取回應錯誤:", err)
		return shoes, err
	}

	// 提取並解析傳回來body.json的資料
	shoes, err = extractSalePageList(body)
	if err != nil {
		fmt.Println("解析 salePageList 錯誤:", err)
		return shoes, err
	}

	// 遍歷訪問shoes.URL，取得每個shoes的Size和Color
	getSizeAndColor(shoes)

	// 篩選出有符合尺寸的鞋子
	filteredShoes := filterShoesBySize(shoes, searchSize)

	return filteredShoes, nil
}

// 提取並解析傳回來body.json的資料，並塞入ListID、Name、Price、Image、URL
func extractSalePageList(body []byte) ([]Shoe, error) {
	var responseData ResponseData
	err := json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, fmt.Errorf("JSON 解析錯誤: %v", err)
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

	for i := range shoes {
		// 增加 WaitGroup 計數
		wg.Add(1)
		go func(i int) {
			// 當 goroutine 完成時減少 WaitGroup 計數
			defer wg.Done()
			// 發送shoes.URL HTTP GET 請求
			childresp, err := http.Get(shoes[i].URL)
			if err != nil {
				fmt.Println("遍歷訪問各商品時請求錯誤:", err)
				return
			}
			defer childresp.Body.Close()

			// 解析 HTML 取得鞋子尺寸與顏色
			size, color, err := extractSizesAndColors(childresp.Body)
			if err != nil {
				fmt.Println("解析 HTML 錯誤:", err)
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

// 解析 HTML 並從中提取鞋子尺寸
func extractSizesAndColors(body io.Reader) ([]string, []string, error) {
	// 讀取 HTML 內容
	htmlBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, nil, fmt.Errorf("讀取 HTML 失敗: %v", err)
	}
	htmlContent := string(htmlBytes)

	// 找出所有 `PropertyNameSet":"尺寸:XX`
	sizeRe := regexp.MustCompile(`"PropertyNameSet"\s*:\s*"[^"]*尺寸:(\d+)"`)
	sizeMatches := sizeRe.FindAllStringSubmatch(htmlContent, -1)

	if sizeMatches == nil {
		return nil, nil, fmt.Errorf("未找到任何尺寸資料")
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
		return nil, nil, fmt.Errorf("未找到任何顏色資料")
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
