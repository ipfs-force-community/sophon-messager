<p align="center">
  <a href="https://damocles.venus-fil.io/" title="Damocles Docs">
    <img src="https://user-images.githubusercontent.com/1591330/205581370-d467d776-60a4-4b37-b25a-58fa82adb156.png" alt="Sophon Logo" width="128" />
  </a>
</p>

<h1 align="center">sophon-messager</h1>

<p align="center">
 <a href="https://github.com/ipfs-force-community/sophon-messager/actions"><img src="https://github.com/ipfs-force-community/sophon-messager/actions/workflows/build_upload.yml/badge.svg"/></a>
 <a href="https://codecov.io/gh/ipfs-force-community/sophon-messager"><img src="https://codecov.io/gh/ipfs-force-community/sophon-messager/branch/master/graph/badge.svg?token=J5QWYWkgHT"/></a>
 <a href="https://goreportcard.com/report/github.com/ipfs-force-community/sophon-messager"><img src="https://goreportcard.com/badge/github.com/ipfs-force-community/sophon-messager"/></a>
 <a href="https://github.com/ipfs-force-community/sophon-messager/tags"><img src="https://img.shields.io/github/v/tag/ipfs-force-community/sophon-messager"/></a>
  <br>
</p>

messager is a component used to manage local messages, with the purpose of saving address messages, managing message status, and controlling the frequency of push messages.

Use [Venus Issues](https://github.com/filecoin-project/venus/issues) for reporting issues about this repository.

### Work

- ✅ Remote wallet support: One messenger support multiple wallets to manage their keys separately
- ✅ Message pool for multiple miners: As a service, Messenger provides API for miners to put messages on chain
- ✅ Supports sqlite local storage and mysql remote storage for more secure and stable storage
- ✅ Scan the address of the miner's wallet, monitor the actor status of address on the chain, maintain the address's nonce information,
- ✅ Fill on fly: gas related parameters and nonce are to be filled out when sending a message on chain according to gas policy, to make sure the gas-estimation and other seeting are valid
- ✅ Maintain message status, including whether the message is chained and replaced. Save the results of the execution.
- ✅ Global Gas estimate paraters, address push quantity configuration.
- ✅ Message-delivery assuring: Auto replace parameters and resend messages whenever there is a failure
- ✅ Multi-point message delivery through daemon program
- ✅ broadcast message through libp2p
- ✅ Enhanced API Security
- 🔲 Rich and flexible message sorting options


### Getting Start

build binary
```sh
git clone https://github.com/ipfs-force-community/sophon-messager.git
make
```

### Set repo path

> The default path is ~/.sophon-messager.
```
./sophon-messager --repo=path_to_repo run
```

### Run

```sh
./sophon-messager run \
--node-url=/ip4/127.0.0.1/tcp/3453 \
--gateway-url=/ip4/127.0.0.1/tcp/45132 \
--auth-url=http://127.0.0.1:8989 \
--auth-token=<auth-token> \
--db-type=sqlite
```

We will find three files in ~/.sophon-messager

* message.db
* message.db-shm
* message.db-wal

#### db use mysql

```sh
./sophon-messager run \
--node-url=/ip4/127.0.0.1/tcp/3453 \
--gateway-url=/ip4/127.0.0.1/tcp/45132 \
--auth-url=http://127.0.0.1:8989 \
--auth-token=<auth-token> \
--db-type=mysql \
--mysql-dsn="user:password@(127.0.0.1:3306)/messager?parseTime=true&loc=Local"
```

### Config

> The configuration file is saved in ~/.sophon-messager/config.toml

```
[api]
  Address = "/ip4/0.0.0.0/tcp/39812"

[db]
  # support sqlite and mysql
  type = "sqlite"

  [db.mysql]
    connMaxLifeTime = "1m0s"
    connectionString = ""
    debug = false
    maxIdleConn = 10
    maxOpenConn = 10

  [db.sqlite]
    debug = false
    file = ""

[gateway]
  # gateway token, generate by auth server
  token = ""
  # gateway url
  url = "/ip4/127.0.0.1/tcp/45132"

  [gateway.cfg]
    RequestQueueSize = 30
    RequestTimeout = "5m0s"

[jwt]
  # auth server url, not connect when empty
  authURL = "http://127.0.0.1:8989"

  [jwt.local]
    # JWT token, generate by secret
    secret = ""
    # hex JWT secret, randam generate first init
    token = ""

[log]
  # default log level
  level = "info"
  # log output file
  path = "messager.log"

[messageService]
  # skip process head
  skipProcessHead = false
  # skip push message
  skipPushMessage = false
  # file used to store tipset
  tipsetFilePath = "./tipset.json"

[messageState]
  CleanupInterval = 86400
  DefaultExpiration = 259200
  backTime = 86400

[node]
  # node token, generate by auth server
  token = ""
  # node url
  url = "/ip4/127.0.0.1/tcp/3453"

[rateLimit]
  redis = "" # eg. 127.0.0.1:6379

[tracing]
  JaegerEndpoint = "" # eg. 1270.0.0.1:6831
  # enable trace
  JaegerTracingEnabled = false
  ProbabilitySampler = 1.0
  ServerName = "sophon-messenger"
```
