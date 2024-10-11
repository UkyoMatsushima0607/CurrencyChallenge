package main

import (
	"currency_challenge/data"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 通貨情報を格納する構造体
type CurrencyInfo struct {
	CODE   string    `json:"CODE"`
	NAME   string    `json:"NAME"`
	RATE   float32   `json:"RATE"`
	DENOMI []float32 `json:"DENOMI"`
	PRICE  float32   `json:"PRICE"`
}

// JSONデータを解析するための構造体
type ApiResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float32 `json:"rates"`
}

func main() {
	r := gin.Default()

	r.Static("/public", "./public")
	r.StaticFile("/favicon.ico", "./public/favicon.ico")

	r.GET("/currency", func(c *gin.Context) {
		// 外部APIからデータを取得
		resp, err := http.Get("https://open.er-api.com/v6/latest/JPY")
		if err != nil {
			log.Printf("Failed to fetch data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch data from the API",
			})
			return
		}
		defer resp.Body.Close()

		// APIレスポンスの内容をデコード
		var apiResponse ApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			log.Printf("Failed to decode API response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to decode API response",
			})
			return
		}

		// 通貨情報を格納するスライス
		var currencyInfos []CurrencyInfo

		// 各通貨について計算を実行
		for code, rate := range apiResponse.Rates {
			name, ok := data.CurrencyUnits[code]
			if !ok {
				continue
			}

			denominations, ok := data.CurrencyDenominations[code]
			if !ok {
				continue
			}

			// 額面とレートを掛け合わせた価格の合計を計算
			var price float32
			for _, denom := range denominations {
				price += rate * denom
			}

			// 通貨情報を構造体に格納
			currencyInfo := CurrencyInfo{
				CODE:   code,
				NAME:   name,
				RATE:   rate,
				DENOMI: denominations,
				PRICE:  price,
			}

			currencyInfos = append(currencyInfos, currencyInfo)
		}

		// ランダムにアクセスするためにシャッフル
		rand.NewSource(time.Now().UnixNano())
		rand.Shuffle(len(currencyInfos), func(i, j int) {
			currencyInfos[i], currencyInfos[j] = currencyInfos[j], currencyInfos[i]
		})

		// 結果をJSONで返却
		c.JSON(http.StatusOK, currencyInfos)
	})

	// サーバーをポート8080で起動
	r.Run(":8080")
}
