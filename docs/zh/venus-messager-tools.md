# venus messager tools

一个基于 sophon-messager 开发的常用工具集

### 帮助

```sh
./sophon-messager-tools -h

NAME:
   sophon-messager-tools - A new cli application

USAGE:
   sophon-messager-tools [global options] command [command options] [arguments...]

COMMANDS:
   batch-replace  batch replace messages
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  The configuration file (default: "./tools_config.toml")
   --help, -h                show help (default: false)
```

### [配置](https://github.com/ipfs-force-community/sophon-messager/blob/main/tools_config.toml)

默认配置是 `tools_config.toml`，使用时需提前把配置配好

```
$cat tools_config.toml

# res
[BatchReplace]
  # 消息在messager阻塞的时间，只包含fillmsg
  BlockTime = "5m"
  # 限定消息的from地址
  From = ""

  [[BatchReplace.Selectors]]
    # 限定消息 from 地址的类型，具体可以参考：https://github.com/filecoin-project/venus/blob/master/venus-shared/builtin-actors/builtin_actors_gen.go
    ActorCode = ""
    # 限定消息的methods
    Methods = [5]

[Messager]
  # sophon-messager 的 token
  Token = ""
  # sophon-messager 的 URL
  URL = "/ip4/127.0.0.1/tcp/39812"

[Venus]
  # 节点的 token
  Token = ""
  # 节点的 URL
  URL = "/ip4/127.0.0.1/tcp/3453"
```

可以使用 `./venus state get-actor <address>` 来查看地址的 actor code

主网不同类型 actor code 如下：

* account: bafk2bzacedudbf7fc5va57t3tmo63snmt3en4iaidv4vo3qlyacbxaa6hlx6y
* cron: bafk2bzacecqb3eolfurehny6yp7tgmapib4ocazo5ilkopjce2c7wc2bcec62
* init: bafk2bzaceaipvjhoxmtofsnv3aj6gj5ida4afdrxa4ewku2hfipdlxpaektlw
* multisig: bafk2bzacebhldfjuy4o5v7amrhp5p2gzv2qo5275jut4adnbyp56fxkwy5fag
* paymentchannel: bafk2bzacebalad3f72wyk7qyilvfjijcwubdspytnyzlrhvn73254gqis44rq
* reward: bafk2bzacecwzzxlgjiavnc3545cqqil3cmq4hgpvfp2crguxy2pl5ybusfsbe
* storagemarket: bafk2bzacediohrxkp2fbsl4yj4jlupjdkgsiwqb4zuezvinhdo2j5hrxco62q
* storageminer: bafk2bzacecgnynvd3tene3bvqoknuspit56canij5bpra6wl4mrq2mxxwriyu
* storagepower: bafk2bzacebjvqva6ppvysn5xpmiqcdfelwbbcxmghx5ww6hr37cgred6dyrpm
* system: bafk2bzacedwq5uppsw7vp55zpj7jdieizirmldceehu6wvombw3ixq2tcq57w
* verifiedregistry: bafk2bzaceb3zbkjz3auizmoln2unmxep7dyfcmsre64vnqfhdyh7rkqfoxlw4


### 批量replace message

根据配置中的条件批量replace长时间阻塞在 sophon-messager 中的消息

> sophon-messager-tools batch-replace [command options] [arguments...]

```
# --max-fee 和 --gas-over-premium 是可选项，gas-over-premium 是 gas premium的系数，该值为0是，则不会起作用
./sophon-messager-tools batch-replace --auto

or

./sophon-messager-tools batch-replace --gas-feecap <value> --gas-premium <value> --gas-limit <value>

```
