# venus-message changelog

## v1.10.0-rc3

1. 升级 venus 和 go-jsonrpc 版本
2. 推送消息接口先解析地址再验证权限

## v1.10.0-rc1

支持 Filecoin NV18 网络升级

* 把 replacedmsg 重命名为 nonceconfict [[#306](https://github.com/filecoin-project/venus-messager/pull/304)]
* 使用 untrust 接口推送消息 [[#306](https://github.com/filecoin-project/venus-messager/pull/306)]
* 按从小到大查询 unchain 消息 [[#307](https://github.com/filecoin-project/venus-messager/pull/307)]
* 移除测试中的重复代码 [[#308](https://github.com/filecoin-project/venus-messager/pull/308)]
