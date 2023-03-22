# venus-message changelog

## v1.10.2

* 修复 list-fail 命令会输出正确的消息 [[#329](https://github.com/filecoin-project/venus-messager/pull/329)]
* ListBlockedMessage 接口也返回 unfill 消息 [[#330](https://github.com/filecoin-project/venus-messager/pull/330)]
* 补充更详细的错误信息 [[#331](https://github.com/filecoin-project/venus-messager/pull/331)]

## v1.10.1

* 支持 delegated 地址的消息 [[#323](https://github.com/filecoin-project/venus-messager/pull/323)]
* 升级 venus 和 venus-auth 版本到 v1.10.1

## v1.10.0

* 升级 venus 和 venus-auth 版本到 v1.10.0
* 升级 go-jsonrpc 版本到 v0.1.7

## v1.10.0-rc3

1. 升级 venus 和 go-jsonrpc 版本
2. 推送消息接口先解析地址再验证权限

## v1.10.0-rc1

支持 Filecoin NV18 网络升级

* 把 replacedmsg 重命名为 nonceconfict [[#306](https://github.com/filecoin-project/venus-messager/pull/304)]
* 使用 untrust 接口推送消息 [[#306](https://github.com/filecoin-project/venus-messager/pull/306)]
* 按从小到大查询 unchain 消息 [[#307](https://github.com/filecoin-project/venus-messager/pull/307)]
* 移除测试中的重复代码 [[#308](https://github.com/filecoin-project/venus-messager/pull/308)]
