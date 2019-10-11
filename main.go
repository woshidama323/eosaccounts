package main
import (
	"github.com/gorilla/websocket"
	"log"
	"time"
	"net/url"
	"encoding/json"
	// "regexp"
	"os"
	"strconv"
	"github.com/go-redis/redis"
	"flag"
)
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init(){
	Info = log.New(os.Stdout,"Info:",log.Ldate | log.Ltime | log.Lshortfile)
}
func main(){

	start := flag.Int("start",1,"起始页")
	stop := flag.Int("stop",1,"终止页")
	minasset := flag.Float64("minasset",1.0,"过滤的最小资产")

	flag.Parse()
	u := url.URL{Scheme:"wss",Host:"api-v1.eosflare.io",Path:"/socket.io/",}
	v := url.Values{}
	v.Add("EIO","3")
	v.Add("transport","websocket")

	urlstr := u.String() + "?" + v.Encode()

	c, _, err := websocket.DefaultDialer.Dial(urlstr, nil)
	if err != nil{
		//链接失败blockchain
		Info.Println("......",err)
	}

	defer c.Close()

	// tc := time.NewTicker(time.Second * 5)
	// go func(){
	// 	i := 0
	// 	for{
	// 		select {
			
	// 		case t := <-tc.C:
	// 			Info.Println("=====",t)
	// 			err := c.WriteMessage(websocket.TextMessage,[]byte("2"))
	// 			if err != nil {
	// 				Info.Println("failed to send the ping message ",err)
	// 			}

	// 			// ["message","{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":1,\"limit\":500,\"sortBy\":\"total\",\"ascending\":false,\"lang\":\"zh-CN\"}"]

	// 			// jsonstr := "{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":" + string(i) + ",\"limit\":500,\"sortBy\":\"total\",\"ascending\":false,\"lang\":\"zh-CN\"}"
	// 			// reqstr := "42" + "[\"message\",\"" + jsonstr + "\"]"
	// 			// err = c.WriteMessage(websocket.TextMessage,[]byte(reqstr))
				
	// 			// if err != nil {
	// 			// 	Info.Println("..........**,.....",err)
	// 			// }
			
	// 		}

	// 		i++
	// 	}
	// }()


	tc2 := time.NewTicker(time.Second * 3)
	go func(){
		i := *start
		Info.Println(".....start....",i)
		for{
			select {
			
			case t := <-tc2.C:
				Info.Println("=====",t)

				// ["message","{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":1,\"limit\":500,\"sortBy\":\"total\",\"ascending\":false,\"lang\":\"zh-CN\"}"]

				jsonstr := "{\\\"_url\\\":\\\"/chain/get_user_whales\\\",\\\"_method\\\":\\\"POST\\\",\\\"_headers\\\":{\\\"content-type\\\":\\\"application/json\\\"},\\\"page\\\":" + strconv.FormatInt(int64(i),10) + ",\\\"limit\\\":500,\\\"sortBy\\\":\\\"total\\\",\\\"ascending\\\":false,\\\"lang\\\":\\\"zh-CN\\\"}"
				reqstr := "42" + "[\"message\",\"" + jsonstr + "\"]"

				Info.Println(reqstr)
				err = c.WriteMessage(websocket.TextMessage,[]byte(reqstr))
				
				if err != nil {
					Info.Println("..........**,.....",err)
				}
			
			}

			i++
			Info.Println(".....monitor....",i)
			Info.Println(".....compare stop....",*stop)
			if i > *stop{
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
		_,msg,err := c.ReadMessage()
		if err != nil {
			Info.Println("failed to get the message..",err)
		}
		// Info.Println(".....++++++++",string(msg))

		var out interface{}

		// if string(msg[:1]) == "3"{
		// 	Info.Println("ping pong msg:",string(msg))
		// 	continue

		// }else 
		if string(msg[:2]) == "42"{
			Info.Println("accounts msg")
			err := json.Unmarshal(msg[2:],&out)
			if err !=nil {
				Info.Println("failed to unmarshal the bytes",err)
			}
		}else {
			continue
		}

		switch m := out.(type){
		case map[string]interface {}:
			// Info.Println("....",m)

			if v,ok := out.(map[string]interface{})["sid"];ok{
				Info.Println("start msg",v)
			}

			// if v2,ok2 := out.(map[string]interface{})["sid"];ok2{

			// 	Info.Println("start msg")
			// }
			case []interface{}:
				// Info.Println("______",out)
				go func(input interface{}){
					// 
					rc := redis.NewClient(&redis.Options{
						Addr:     "localhost:6379",
						Password: "", // no password set
						DB:       14,//14,  // use default DB
					})

					getholder := input.([]interface{})[1].(string)
					Info.Println(getholder)

					var test interface{}
					json.Unmarshal([]byte(getholder),&test)
					for _,x := range test.(map[string]interface{})["holders"].([]interface{}){

						
						liquid := x.(map[string]interface{})["liquidity"].(float64)
						if liquid < *minasset{
							Info.Println("it's the small account: ",x)
							continue
						}
						err = rc.SAdd("forloopsend",x.(map[string]interface{})["owner"].(string)).Err()
						if err != nil {
							Info.Println("get errors ??...",err)
							panic(err)
						}
					}

				}(out)

		default:
			Info.Println("Unsupported message", m)
		}
	}
}