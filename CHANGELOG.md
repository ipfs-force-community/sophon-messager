# venus-message changelog

## v1.11.0

* bump up version to v1.11.0

## v1.11.0-rc1

### Features
* feat: support method type fee cfg / 支持消息类型级别的费用配置  [[#303](https://github.com/ipfs-force-community/sophon-messager/pull/303)]
* feat: add status api to detect api ready p /添加状态检测接口 [[#313](https://github.com/ipfs-force-community/sophon-messager/pull/313)]
* feat: update the authClient with token  /客户端token验证 [[#317](https://github.com/ipfs-force-community/sophon-messager/pull/317)]
* chore: more detailed error information /更加详细的错误信息 [[#331](https://github.com/ipfs-force-community/sophon-messager/pull/331)]
* feat: ListBlockedMessage interface also returns unfill message  / 同样返回Unfill的消息 [[#330](https://github.com/ipfs-force-community/sophon-messager/pull/330)]
* feat: add docker push  / 增加推送到镜像仓库的功能 [[#335](https://github.com/ipfs-force-community/sophon-messager/pull/335)]
* feat: Reduce minimum replace fee from 1.25x to 1.1x  / 最小 replace fee 乘数改为 1.1x [[#336](https://github.com/ipfs-force-community/sophon-messager/pull/336)]


### Bug Fixes
* fix: No actor configuration was used when replacing messages  /replace 消息时没有使用 actor config 表中的 maxfee[[#328](https://github.com/ipfs-force-community/sophon-messager/pull/328)]
* fix: failed create actor_cfg table in mysql  / 修复创建actor_cfg表失败[[#327](https://github.com/ipfs-force-community/sophon-messager/pull/327)]
* fix: Exclude empty strings when listing failure message /排除 error_msg 为空的时候被认为是 failed 消息[[#329](https://github.com/ipfs-force-community/sophon-messager/pull/329)]
* fix: Modify the WaitingChainHeadStableDuration value only for 2k networks  / 2k网络中修改WaitingChainHeadStableDuration的值[[#334](https://github.com/ipfs-force-community/sophon-messager/pull/334)]

## v1.10.1

* 支持 delegated 地址的消息 [[#323](https://github.com/ipfs-force-community/sophon-messager/pull/323)]
* 升级 venus 和 venus-auth 版本到 v1.10.1

## v1.10.0

* 升级 venus 和 venus-auth 版本到 v1.10.0
* 升级 go-jsonrpc 版本到 v0.1.7

## v1.10.0-rc3

1. 升级 venus 和 go-jsonrpc 版本
2. 推送消息接口先解析地址再验证权限

## v1.10.0-rc1

支持 Filecoin NV18 网络升级

* 把 replacedmsg 重命名为 nonceconfict [[#306](https://github.com/ipfs-force-community/sophon-messager/pull/304)]
* 使用 untrust 接口推送消息 [[#306](https://github.com/ipfs-force-community/sophon-messager/pull/306)]
* 按从小到大查询 unchain 消息 [[#307](https://github.com/ipfs-force-community/sophon-messager/pull/307)]
* 移除测试中的重复代码 [[#308](https://github.com/ipfs-force-community/sophon-messager/pull/308)]
