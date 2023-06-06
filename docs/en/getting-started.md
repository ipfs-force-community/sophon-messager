# How to use venus messager

messager is a component used to manage local messages, with the purpose of saving address messages, managing message status, and controlling the frequency of push messages.

## Getting start

### Clone this git repository to your machine

```bash
git clone https://github.com/ipfs-force-community/sophon-messager.git
```

### Install Dependencies and Build

```bash
make
```

### Run messager

```bash
./sophon-messager run [options]
```

```bash
options:
  --auth-url           url for auth server
  --auth-token         token for auth server
  --node-url           url for connection lotus/venus
  --node-token         token auth for lotus/venus
  --db-type            which db to use. sqlite/mysql
  --mysql-dsn          mysql connection string
  --gateway-url        url for gateway server
  --gateway-token      token for gateway server
  --rate-limit-redis   limit flow using redis
```

## Commands

### Message commands

1. search message

```bash
./sophon-messager msg search --id=<message id>
```

2. list message

```bash
./sophon-messager msg list
# list messages with the same address
./sophon-messager msg list --from <address>
```

3. update one filled message state

```bash
./sophon-messager msg update_filled_msg --id=<message id>
```

4. update all filled message state

```bash
./sophon-messager msg update_all_filled_msg
```

5. wait a message result by id

```bash
./sophon-messager msg wait <message id>
```

6. republish a message by id

```bash
./sophon-messager msg republish <message id>
```

7. replace a message

```bash
./sophon-messager msg replace --gas-feecap=[gas-feecap] --gas-premium=[gas-premium] --gas-limit=[gas-limit] --auto=[auto] --max-fee=[max-fee] <message-id>
# or
./sophon-messager msg replace --gas-feecap=[gas-feecap] --gas-premium=[gas-premium] --gas-limit=[gas-limit] --auto=[auto] --max-fee=[max-fee] <from> <nonce>
```

8. list failed messages, maybe signed message failed or gas estimate failed

```bash
./sophon-messager msg list-fail
```

9. lists message that have not been chained for a period of time

```bash
./sophon-messager msg list-blocked
```

10. manual mark error messages

```bash
./sophon-messager msg mark-bad <message id>
```

### Address commands

1. search address

```bash
./sophon-messager address search <address>
```

2. list address

```bash
./sophon-messager address list
```

3. reset address

> The nonce of the address is set to nonce on the chain, and all unchain messages are marked as failed messages

```bash
./sophon-messager reset <address>
```

4. forbidden address

```bash
./sophon-messager address forbidden <address>
```

5. activate a frozen address

```bash
./sophon-messager address active <address>
```

6. set the number of address selection messages

```bash
./sophon-messager address set-sel-msg-num --num=5 <address>
```

7. set parameters related to address fee

> sophon-messager address set-fee-params [options] address

```bash
 # options
 # --gas-overestimation value  Estimate the coefficient of gas (default: 0)
 # --gas-feecap value          Gas feecap for a message (burn and pay to miner, attoFIL/GasUnit)
 # --max-fee value             Spend up to X attoFIL for message
 # --gas-over-premium value    Coefficient of gas premium (default: 0)

./sophon-messager address set-fee-params <address>
```

### shared params commands

1. get shared params

```bash
./sophon-messager share-params get
```

2. set shared params

```bash
./sophon-messager share-params set --gas-over-estimation=1.25 --gas-feecap="0" --max-fee="7000000000000000" --sel-msg-num=20 --gas-over-premium 1
```

3. manual refresh shared params from DB

```bash
./sophon-messager share-params refresh
```

### node commands

1. search node info by name

```bash
./sophon-messager node search <name>
```

2. add node info

```bash
./sophon-messager node add --name=<node-name> --url=<node-url> --token=<node-token>
```

3. list node info

```bash
./sophon-messager node list
```

4. del node info by name

```bash
./sophon-messager node del <name>
```

### log

1. set log level

```bash
# eg. trace,debug,info,warn|warning,error,fatal,panic
./sophon-messager log set-level
```

### send 命令

> send message
> sophon-messager send [command options] [targetAddress] [amount]

```
   options:
   --from value         optionally specify the address to send
   --gas-premium value  specify gas price to use in AttoFIL (default: "0")
   --gas-feecap value   specify gas fee cap to use in AttoFIL (default: "0")
   --gas-limit value    specify gas limit (default: 0)
   --method value       specify method to invoke (default: 0)
   --params-json value  specify invocation parameters in json
   --params-hex value   specify invocation parameters in hex
```
