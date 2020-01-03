## method for using
dlv debug -- -start=1 -stop=1 -minasset=0.01 -contract=athenastoken -symbol=ATHENA -rediskey=testrediskey


42["message","{\"_url\":\"/chain/get_user_whales\",\"_method\":\"POST\",\"_headers\":{\"content-type\":\"application/json\"},\"page\":0,\"limit\":50,\"sortBy\":\"liquidity\",\"ascending\":false,\"lang\":\"zh-CN\"}"]



./testadv -start=358 -stop=1000 -minasset=0.01 -rediskey=new50wfilter



## 动态获取eosflares上的数据，以newdexpublic账户中send类型来过滤

1. 设计方案，
cron定时任务，每分钟调用一次，
每次获取第一页的数据
每夜数据500条
存入到redis中，sadd方式，不会重复 

### 定时任务
crontab -e 




e.g.
dlv debug -- -start=0 -stop=1000 -contract=newdexpublic



