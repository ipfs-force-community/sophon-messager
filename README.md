# Venus-Messenger
Venus-Messenger is a component of Venus for managing messages. It can also work as
an independent service to manage message pool and message selection for any other
implementations.

## Major Features:

- Message pool for multiple miners: As a service, Messenger provides API for miners to put messages on chain
- Gas policy management: Gas policy can be configured at address level, each address can have
different gas policy, including fee-cap, premium, etc.
- Fill on fly: gas related parameters and nonce are to be filled out when sending a message on
chain according to gas policy, to make sure the gas-estimation and other seeting are valid
- message-delivery assuring: Auto replace parameters and resend messages whenever there is a failure
- Maximum message selection for multiple miners: Whenever there are multiple chances to generate block in a particular epoch when serving multiple miners, the smart message selection could avoid redundant message packaging
- Remote wallet support: One messenger support multiple wallets to manage their keys separately
- Message prioritization: Since nonce is set on the fly, the message prioritization could be flexible, e.g. make sure ProveCommit message to be sent in x hours, etc.

