# 配置文件

默认配置文件是～/.sophon-messager/config.toml。`#`表示是注释。

```toml
[api]
  Address = "/ip4/127.0.0.1/tcp/39812"  #messager的监听地址

[db]
  type = "sqlite"  #数据库类型。mysql或者sqlite

  [db.mysql]
    connMaxLifeTime = "1m0s"
    connectionString = ""
    debug = false
    maxIdleConn = 10
    maxOpenConn = 10

  [db.sqlite]
    debug = false

[gateway]
  token = ""   #[gateway],[jwt],[node]三个字段基本上都是用同一个auth服务的token
  url = ["/ip4/127.0.0.1/tcp/45132"]

[jwt]
  authURL = "http://127.0.0.1:8989"
  token = "" #[gateway],[jwt],[node]三个字段基本上都是用同一个auth服务的token

# messager直接通过p2p给链节点（venus/lotus）发送消息
# 可选
[libp2p]
  bootstrapAddresses = []
  expandPeriod = "0s"
  listenAddresses = "/ip4/0.0.0.0/tcp/0"
  minPeerThreshold = 0

[log]
  level = "info"
  path = ""

[messageService]
  DefaultTimeout = "1s"  #请求链节点接口的超时时长 
  EstimateMessageTimeout = "5s" #调用链节点进行消息预估gas费等的超时时长
  SignMessageTimeout = "3s" #调用gateway请求wallet进行签名的超时时长
  WaitingChainHeadStableDuration = "8s" #messager收到一个newhead消息后，如果8秒内没有收到新的newhead，就会认为收到的newhead是stable的了
  skipProcessHead = false #是否更新消息上链后的状态。在多个messager共用一个数据库时，只需要一个messager进行消息的全部状态更新
  skipPushMessage = false  #不推送消息到链。在多个messager共用一个数据库时，不推送消息的messager只做接受消息的任务，另外的messager进行推送消息

[metrics]
  Enabled = false

  [metrics.Exporter]
    Type = "prometheus"

    [metrics.Exporter.Graphite]
      Host = "127.0.0.1"
      Namespace = ""
      Port = 4568
      ReportingPeriod = "10s"

    [metrics.Exporter.Prometheus]
      EndPoint = "/ip4/0.0.0.0/tcp/4568"
      Namespace = ""
      Path = "/debug/metrics"
      RegistryType = "define"
      ReportingPeriod = "10s"

#venus同步节点或者sophon-co负载代理服务
[node]
  token = "" #[gateway],[jwt],[node]三个字段基本上都是用同一个auth服务的token
  url = "/ip4/127.0.0.1/tcp/3453"

[publisher]
  cacheReleasePeriod = 0 #间隔多久缓存清理一次
  concurrency = 5 #同时推送消息的线程数
  enableMultiNode = true #是否可以给多个节点推送消息
  enablePubsub = false #是否通过p2p网络发送消息。需要和[libp2p]配置一起使用

#可选
[rateLimit]
  redis = ""

#可选
[tracing]
  JaegerEndpoint = "localhost:6831"
  JaegerTracingEnabled = false
  ProbabilitySampler = 1.0
  ServerName = ""

```