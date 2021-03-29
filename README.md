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
- ❓ Manage messages through a multi-tenant pattern by wallet name


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