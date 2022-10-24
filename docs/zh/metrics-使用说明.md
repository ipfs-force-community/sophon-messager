# metrics 使用说明

## 配置

`Metrics` 基本的配置样例如下：
```toml
[Metrics]
  # 是否开启metrics指标统计，默认为false
  Enabled = false
  
  [Metrics.Exporter]
    # 指标导出器类型，目前可选：prometheus或graphite，默认为prometheus
    Type = "prometheus"
    
    [Metrics.Exporter.Prometheus]
      # multiaddr
      EndPoint = "/ip4/0.0.0.0/tcp/4568"
      # 命名规范: "a_b_c", 不能带"-"
      Namespace = "messager01" 
      # 指标注册表类型，可选：default（默认，会附带程序运行的环境指标）或 define（自定义）
      RegistryType = "define"
      # prometheus 服务路径
      Path = "/debug/metrics"
      # 上报周期
      ReportingPeriod = "10s"
      
    [Metrics.Exporter.Graphite]
      # 命名规范: "a_b_c", 不能带"-"
      Namespace = "messager01" 
      # graphite exporter 收集器服务地址
      Host = "127.0.0.1"
      # graphite exporter 收集器服务监听端口
      Port = 4568
      # 上报周期
      ReportingPeriod = "10s"
```
## 导出器

目前可以选择两类导出器（`exporter`）：`Prometheus exporter` 或 `Graphite exporter`，默认是前者。

`exporter` 端口为 `4568`，url为 `debug/metrics`, 因此对于默认的部署方式，`exporter` 的url为 `host:4568/debug/metrics`

如果配置 `Prometheus exporter`，则在 `venus-messager` 服务启动时会附带启动 `Prometheus exporter` 的监听服务，可以通过以下方式快速查看指标：


```bash
 $  curl http://localhost:4568/debug/metrics
   # HELP messager01_chain_head_stable_dur_s Duration of chain head stabilization
   # TYPE messager01_chain_head_stable_dur_s histogram
   messager01_chain_head_stable_dur_s_bucket{le="8"} 0
   messager01_chain_head_stable_dur_s_bucket{le="9"} 11
   messager01_chain_head_stable_dur_s_bucket{le="10"} 27
   messager01_chain_head_stable_dur_s_bucket{le="12"} 43
   messager01_chain_head_stable_dur_s_bucket{le="14"} 48
   messager01_chain_head_stable_dur_s_bucket{le="16"} 49
   messager01_chain_head_stable_dur_s_bucket{le="18"} 49
   messager01_chain_head_stable_dur_s_bucket{le="20"} 49
   messager01_chain_head_stable_dur_s_bucket{le="25"} 49
   messager01_chain_head_stable_dur_s_bucket{le="30"} 49
   messager01_chain_head_stable_dur_s_bucket{le="60"} 50
   messager01_chain_head_stable_dur_s_bucket{le="+Inf"} 50
   messager01_chain_head_stable_dur_s_sum 503.99999999999983
   messager01_chain_head_stable_dur_s_count 50
   # HELP messager01_chain_head_stable_s Delay of chain head stabilization
   # TYPE messager01_chain_head_stable_s gauge
   messager01_chain_head_stable_s 9
   ... ...
```
> 如果遇到错误 `curl: (56) Recv failure: Connection reset by peer`, 请使用本机 `ip` 地址, 如下所示:
```bash
$  curl http://<ip>:4568/debug/metrics
```

如果配置 `Graphite exporter`，需要先启动 `Graphite exporter` 的收集器服务， `venus-messager` 服务启动时将指标上报给收集器。服务启动参考 [Graphite exporter](https://github.com/prometheus/graphite_exporter) 中的说明。

`Graphite exporter` 和 `Prometheus exporter` 自身都不带图形界面的，如果需要可视化监控及更高阶的图表分析，请到 `venus-docs` 项目中查找关于 `Prometheus+Grafana` 的说明文档。

## 指标

### 地址

```
# 地址余额
WalletBalance    = stats.Int64("wallet_balance", "Wallet balance", stats.UnitDimensionless)
# 地址在数据库中的nonce值
WalletDBNonce    = stats.Int64("wallet_db_nonce", "Wallet nonce in db", stats.UnitDimensionless)
# 地址链上nonce值
WalletChainNonce = stats.Int64("wallet_chain_nonce", "Wallet nonce on the chain", stats.UnitDimensionless)
```

### 消息数量

```
# unfill消息数量，可以根据地址分组
NumOfUnFillMsg = stats.Int64("num_of_unfill_msg", "The number of unFill msg", stats.UnitDimensionless)
# fill消息数量，可以根据地址分组
NumOfFillMsg   = stats.Int64("num_of_fill_msg", "The number of fill Msg", stats.UnitDimensionless)
# failed消息数量
NumOfFailedMsg = stats.Int64("num_of_failed_msg", "The number of failed msg", stats.UnitDimensionless)

# fill消息三分未上链的数量
NumOfMsgBlockedThreeMinutes = stats.Int64("blocked_three_minutes_msgs", "Number of messages blocked for more than 3 minutes", stats.UnitDimensionless)
# fill消息五分组未上链的数量
NumOfMsgBlockedFiveMinutes  = stats.Int64("blocked_five_minutes_msgs", "Number of messages blocked for more than 5 minutes", stats.UnitDimensionless)
```

### 单次选择消息情况

```
# 选择的消息数量
SelectedMsgNumOfLastRound = stats.Int64("selected_msg_num", "Number of selected messages in the last round", stats.UnitDimensionless)
# 还未上链的fill消息
ToPushMsgNumOfLastRound   = stats.Int64("topush_msg_num", "Number of to-push messages in the last round", stats.UnitDimensionless)
# 过期的消息数量
ExpiredMsgNumOfLastRound  = stats.Int64("expired_msg_num", "Number of expired messages in the last round", stats.UnitDimensionless)
# 错误的消息数量
ErrMsgNumOfLastRound      = stats.Int64("err_msg_num", "Number of err messages in the last round", stats.UnitDimensionless)
```

### head

```
# 链head稳定的花费时间
ChainHeadStableDelay    = stats.Int64("chain_head_stable_s", "Delay of chain head stabilization", stats.UnitSeconds)
```
