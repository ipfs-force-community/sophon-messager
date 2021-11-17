# venus-messager

messager is a component used to manage local messages, with the purpose of saving address messages, managing message status, and controlling the frequency of push messages.

## Work

- ✅ Remote wallet support: One messenger support multiple wallets to manage their keys separately
- ✅ Message pool for multiple miners: As a service, Messenger provides API for miners to put messages on chain
- ✅ Supports sqlite local storage and mysql remote storage for more secure and stable storage
- ✅ Scan the address of the miner's wallet, monitor the actor status of address on the chain, maintain the address's nonce information,
- ✅ Fill on fly: gas related parameters and nonce are to be filled out when sending a message on chain according to gas policy, to make sure the gas-estimation and other seeting are valid
- ✅ Maintain message status, including whether the message is chained and replaced. Save the results of the execution.
- 🚧 Global Gas estimate paraters, address push quantity configuration.
- 🚧 Multi-point message delivery (directly to the blockchain network with libp2p, push to the node by Mpool API), to make sure that messages are propagation over the network
- 🔲 Enhanced API Security
- 🔲 Rich and flexible message sorting options
- 🔲 Message-delivery assuring: Auto replace parameters and resend messages whenever there is a failure

## Getting Start

build binary
```sh
git clone https://github.com/filecoin-project/venus-messager.git
make
```

## Config

```
[api]
  Address = "/ip4/0.0.0.0/tcp/39812"

[db]
  # support sqlite and mysql
  type = "sqlite"

  [db.mysql]
    connMaxLifeTime = "1m0s"
    connectionString = "" # eg. root:password@(127.0.0.1:3306)/messager?parseTime=true&loc=Local
    debug = false
    maxIdleConn = 10
    maxOpenConn = 10

  [db.sqlite]
    debug = false
    file = "./message.db"

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
  ServerName = "venus-messenger"
```

## Run

```sh
./venus-messager run --auth-url=http://127.0.0.1:8989 --node-url=/ip4/127.0.0.1/tcp/3453 --gateway-url=/ip4/127.0.0.1/tcp/45132 --auth-token=<auth-token>
```