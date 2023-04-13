/*
Copyright © 2023 laoyutang

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var apiKey, proxyUrl string

const limitUrl = "https://api.openai.com/dashboard/billing/subscription"
const usageUrl = "https://api.openai.com/dashboard/billing/usage"

var rootCmd = &cobra.Command{
	Use:   "openai-balance",
	Short: "获取openai余额",
	Long: `获取openai余额
	-k 指定api key(必选)
	-p 指定网络代理地址
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// apikey必填
		if len(apiKey) == 0 {
			fmt.Println("请使用-k传入apikey!")
			os.Exit(1)
		}

		var client *http.Client

		// 如果传入了代理地址
		if len(proxyUrl) > 0 {
			proxyUri, parseErr := url.Parse(proxyUrl)
			if parseErr != nil {
				fmt.Println("代理地址解析错误：", parseErr)
				os.Exit(1)
			}
			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUri),
				},
			}
		} else {
			client = http.DefaultClient
		}

		type Query struct {
			startTime time.Time
			endTime   time.Time
		}

		resChan := make(chan *map[string]interface{}, 4)

		getRequest := func(url string, params ...any) {
			request, createErr := http.NewRequest("GET", url, nil)
			if createErr != nil {
				fmt.Println("请求创建失败：", createErr)
				os.Exit(1)
			}

			request.Header.Add("Authorization", "Bearer "+apiKey)

			if url == usageUrl {
				query := request.URL.Query()
				query.Add("start_date", params[0].(Query).startTime.Format("2006-01-02"))
				query.Add("end_date", params[0].(Query).endTime.Format("2006-01-02"))
				request.URL.RawQuery = query.Encode()
			}

			response, resErr := client.Do(request)
			if resErr != nil {
				fmt.Println("请求失败：", resErr)
				os.Exit(1)
			}

			defer response.Body.Close()

			resBytes, readErr := io.ReadAll(response.Body)
			if readErr != nil {
				fmt.Println("读入body错误:", readErr)
				os.Exit(1)
			}

			res := &map[string]interface{}{}
			if jsonErr := json.Unmarshal(resBytes, res); jsonErr != nil {
				fmt.Println("json解析错误:", jsonErr)
				os.Exit(1)
			}

			resChan <- res
		}

		// 获取hardLimit
		go getRequest(limitUrl)
		limit := <-resChan
		hardLimit := (*limit)["hard_limit_usd"].(float64)

		// 获取使用量
		startTime, _ := time.Parse("2006-01-02", "2023-01-01")
		endTime := startTime.AddDate(0, 3, 0)
		count := 0
		var usage float64 = 0

		for startTime.Before(time.Now()) {
			count++
			go getRequest(usageUrl, Query{startTime, endTime})
			startTime = endTime
			endTime = endTime.AddDate(0, 3, 0)
		}

		for {
			res := <-resChan
			usage += (*res)["total_usage"].(float64)
			if count--; count == 0 {
				close(resChan)
				break
			}
		}

		usage = math.Ceil(usage)

		fmt.Println(hardLimit - usage/100)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&apiKey, "key", "k", "", "Openai的apikey")
	rootCmd.Flags().StringVarP(&proxyUrl, "proxy", "p", "", "使用的http代理地址")
}
