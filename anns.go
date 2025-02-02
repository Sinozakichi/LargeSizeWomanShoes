package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
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

	// 驗證篩選出的資料
	for _, shoe := range shoes {
		fmt.Printf("ListID:%s, Name: %s, Price: %s,Image: %s,URL:%s", shoe.ListID, shoe.Name, shoe.Price, shoe.Image, shoe.URL)
	}

	return shoes, nil
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
	for i := range shoes {
		childresp, err := http.Get(shoes[i].URL)
		if err != nil {
			fmt.Println("遍歷訪問各商品時請求錯誤:", err)
			return
		}
		defer childresp.Body.Close()

		// 解析 HTML 取得鞋子尺寸
		size, err := extractSizes(childresp.Body)
		if err != nil {
			fmt.Println("解析 HTML 錯誤:", err)
			continue
		}
		shoes[i].Size = size
		// 解析 HTML 取得鞋子顏色
		color := extractColors(childresp.Body)
		if err != nil {
			fmt.Println("解析 HTML 錯誤:", err)
			continue
		}
		shoes[i].Color = color

	}
}

// 解析 HTML 並從中提取鞋子尺寸
func extractSizes(body io.Reader) ([]string, error) {
	var sizes []string
	tokenizer := html.NewTokenizer(body)
	inSKUList := false

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return sizes, nil // 讀取完畢
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "ul" {
				// 檢查是否為 class="sku-ul"
				for _, attr := range token.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "sku-ul") {
						inSKUList = true
					}
				}
			} else if inSKUList && token.Data == "a" {
				// 提取 <a> 內的文字
				tt = tokenizer.Next()
				if tt == html.TextToken {
					sizes = append(sizes, strings.TrimSpace(tokenizer.Token().Data))
				}
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			if token.Data == "ul" {
				inSKUList = false
			}
		}
	}
}

// 解析 HTML 並從中提取鞋子顏色
func extractColors(body io.Reader) []string {
	doc, err := html.Parse(body)
	if err != nil {
		fmt.Println("解析 HTML 失敗:", err)
		return nil
	}

	var colors []string
	var findUl bool

	// 遞迴遍歷節點
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "ul" {
			// 檢查 ul 是否符合 class="group-list is-circle"
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "group-list is-circle") {
					findUl = true
				}
			}
		}

		if findUl && n.Type == html.ElementNode && n.Data == "p" {
			// 檢查 class 是否為 group-list-item__tooltip_content
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "group-list-item__tooltip_content" {
					if n.FirstChild != nil {
						colors = append(colors, n.FirstChild.Data)
					}
				}
			}
		}

		// 遞迴處理子節點
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return colors
}
