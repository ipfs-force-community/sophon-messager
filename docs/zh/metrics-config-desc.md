# metrics 配置及使用说明

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
      # 上报周期，单位为 秒（s）
      ReportingPeriod = 10
      
    [Metrics.Exporter.Graphite]
      # 命名规范: "a_b_c", 不能带"-"
      Namespace = "messager01" 
      # graphite exporter 收集器服务地址
      Host = "127.0.0.1"
      # graphite exporter 收集器服务监听端口
      Port = 4568
      # 上报周期，单位为 秒（s）
      ReportingPeriod = 10
```
## 导出器

目前可以选择两类导出器（`exporter`）：`Prometheus exporter` 或 `Graphite exporter`，默认是前者。

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
