package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	// "regexp"
	"flag"
	"os"
	"strconv"

	"github.com/go-redis/redis"
)

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
	minasset := flag.Float64("minasset", 1.0, "过滤的最小资产")
	con := flag.String("contract", "", "checking for a contract token")
	// symbol := flag.String("symbol", "", "the token name")

	rediskey := flag.String("rediskey", "forloopsend", "a key for storing the msg to redis")

	flag.Parse()
	u := url.URL{Scheme: "wss", Host: "api-v1.eosflare.io", Path: "/socket.io/"}
	v := url.Values{}
	v.Add("EIO", "3")
	v.Add("transport", "websocket")

	urlstr := u.String() + "?" + v.Encode()

	he := http.Header{}
	he.Add("origin", "https://eosflare.io")
	c, _, err := websocket.DefaultDialer.Dial(urlstr, he)
	if err != nil {
		//链接失败blockchain
		Info.Println("......", err)
		return
	}

	defer c.Close()

	tc2 := time.NewTicker(time.Second * 3)
	go func() {
		i := *start
		Info.Println(".....start....", i)
		for {
			select {

			case t := <-tc2.C:
				Info.Println("=====", t)

				// ["message","{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":1,\"limit\":500,\"sortBy\":\"total\",\"ascending\":false,\"lang\":\"zh-CN\"}"]

				jsonstr := get_user_whales + strconv.FormatInt(int64(i), 10) + "," + "\\\"account\\\":\\\"" + *con + "\\\"" + tokenurl

				reqstr := "42" + "[\"message\",\"" + jsonstr + "\"]"

				Info.Println("final.request..", reqstr)
				err = c.WriteMessage(websocket.TextMessage, []byte(reqstr))

				if err != nil {
					Info.Println("..........**,.....", err)
					time.Sleep(1)
					continue
				}

			}

			i++
			Info.Println(".....monitor....", i)
			Info.Println(".....compare stop....", *stop)
			if i > *stop {
				Info.Println("have gotten the 25w users.")
				break
			}
			// break
		}
	}()

	// go func(){

	// 	// ["message","{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":1,\"limit\":500,\"sortBy\":\"total\",\"ascending\":false,\"lang\":\"zh-CN\"}"]
	// 	c.WriteMessage(websocket.TextMessage,)
	// }()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			Info.Println("failed to get the message..", err)
			//如果链接中断，重新链接
			c, _, err = websocket.DefaultDialer.Dial(urlstr, he)
			if err != nil {
				//链接失败blockchain
				Info.Println("......", err)
				time.Sleep(1)
				continue
			}

		}
		Info.Println(".....++++++++", string(msg))

		var out interface{}

		if string(msg[:1]) == "0" {
			Info.Println("ping pong msg:", string(msg))
			continue
		}

		if len(msg) < 2 {
			continue
		}
		if string(msg[:2]) == "42" {
			Info.Println("accounts msg")
			err := json.Unmarshal(msg[2:], &out)
			if err != nil {
				Info.Println("failed to unmarshal the bytes", err)
			}
		} else {
			continue
		}

		switch m := out.(type) {
		case map[string]interface{}:
			// Info.Println("....",m)

			if v, ok := out.(map[string]interface{})["sid"]; ok {
				Info.Println("start msg", v)
			}

		// if v2,ok2 := out.(map[string]interface{})["sid"];ok2{

		// 	Info.Println("start msg")
		// }
		case []interface{}:
			// Info.Println("______",out)
			go func(input interface{}) {
				//
				rc := redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "", // no password set
					DB:       14, //14,  // use default DB
				})

				getholder := input.([]interface{})[1].(string)
				// Info.Println(getholder)

				var test interface{}
				json.Unmarshal([]byte(getholder), &test)
				if test.(map[string]interface{})["holders"] == nil {
					Info.Println("something wrong with it ..", test)
					return
				}
				for _, x := range test.(map[string]interface{})["holders"].([]interface{}) {

					liquid := 1.0
					if *con != "" {
						liquid = x.(map[string]interface{})["balance"].(float64)
					} else {
						liquid = x.(map[string]interface{})["liquidity"].(float64)
					}

					if liquid < *minasset {
						Info.Println("it's the small account: ", x)
						continue
					}
					storestr := x.(map[string]interface{})["owner"].(string) + "_" + strconv.FormatFloat(liquid, 'f', -1, 64)
					Info.Println("....._+_+_+", storestr)
					err = rc.SAdd(*rediskey, storestr).Err()
					if err != nil {
						Info.Println("get errors ??...", err)
						panic(err)
					}
				}

			}(out)

		default:
			Info.Println("Unsupported message", m)
		}
	}
}
