package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"flag"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"strings"

	"golang.org/x/net/html"
	"github.com/gorilla/websocket"
	"strings"

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	Info = log.New(os.Stdout, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
}

var (
	get_user_whales            = "{\\\"_url\\\":\\\"/chain/get_account_actions\\\",\\\"_method\\\":\\\"POST\\\",\\\"_headers\\\":{\\\"content-type\\\":\\\"application/json\\\"},\\\"page\\\":"
	get_user_whales_with_token = "{\\\"_url\\\":\\\"/chain/get_token_holder_ranks\\\",\\\"_method\\\":\\\"POST\\\",\\\"_headers\\\":{\\\"content-type\\\":\\\"application/json\\\"},\\\"page\\\":"
	commonurl                  = ",\\\"limit\\\":500,\\\"sortBy\\\":\\\"liquidity\\\",\\\"ascending\\\":false,\\\"lang\\\":\\\"zh-CN\\\"}"
	tokenurl                   = ",\\\"contract_account\\\":\\\"\\\",\\\"contract_name\\\":\\\"\\\",\\\"filterSpam\\\":true,\\\"limit\\\":50,\\\"lang\\\":\\\"zh-CN\\\"}"
)

// \"account\":\"newdexpublic\",\"contract_account\":\"\",\"contract_name\":\"\",\"filterSpam\":true,\"page\":0,\"limit\":500,\"lang\":\"zh-CN\"}"]

func main() {

	start := flag.Int("start", 1, "起始页")
	stop := flag.Int("stop", 1, "终止页")
	// minasset := flag.Float64("minasset", 1.0, "过滤的最小资产")
	con := flag.String("contract", "", "checking for a contract token")
	// symbol := flag.String("symbol", "", "the token name")

	rediskey := flag.String("rediskey", "forloopsend", "a key for storing the msg to redis")

	flag.Parse()

	u := url.URL{Scheme: "https", Host: "api.newdex.vip", Path: "/v1/candles"}
	v := url.Values{}

	//https://api.newdex.vip/v1/candles?symbol=defindinvest-defi-eos&time_frame=1day&size=100
	v.Add("symbol", "defindinvest-defi-eos")
	v.Add("time_frame", "1day")
	v.Add("size",100)

	urlstr := u.String() + "?" + v.Encode()

	resp, err := http.Get(urlstr)
	if err != nil {
		//链接失败blockchain
		Error.Println("......", err)
		return
	}

	defer resp.Body.Close()

	
	
}
