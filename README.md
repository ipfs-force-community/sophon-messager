# venus-messager

messager is a component used to manage local messages, with the purpose of saving address messages, managing message status, and controlling the frequency of push messages.

## Work

✅Support for sqlite and mysql storage
✅Connect multiple wallets and scan wallet addresses.
✅Connect node components to maintain the status of messages that have been sent.
✅maintain the status of address(nonce)
✅Simple message selection push and nonce assignment
❌global Gas evaluation parameters, address push quantity configuration
❌Message multi-point push (pushed directly to the blockchain network with libp2p, push to the node by Mpool API)
❌API Security
❌Rich and flexible message sorting options
❓Manage messages through a multi-tenant pattern by wallet name


## Getting Start

build binary
```sh
git clone 
make deps
make
```

edit messager.toml config file, edit node url and token

```sh
./venus-messager -config ./messager.toml
```

## client

