package config

import (
	"time"

	"github.com/ipfs-force-community/metrics"
)

type Config struct {
	DB              DbConfig               `toml:"db"`
	JWT             JWTConfig              `toml:"jwt"`
	Log             LogConfig              `toml:"log"`
	API             APIConfig              `toml:"api"`
	Node            NodeConfig             `toml:"node"`
	MessageService  MessageServiceConfig   `toml:"messageService"`
	Gateway         GatewayConfig          `toml:"gateway"`
	RateLimit       RateLimitConfig        `toml:"rateLimit"`
	Trace           *metrics.TraceConfig   `toml:"tracing"`
	Metrics         *metrics.MetricsConfig `toml:"metrics"`
	Libp2pNetConfig *Libp2pNetConfig       `toml:"libp2p"`
}

type NodeConfig struct {
	Url   string `toml:"url"`
	Token string `toml:"token"`
}

type LogConfig struct {
	Path  string `toml:"path"`
	Level string `toml:"level"`
}

type APIConfig struct {
	Address string
}

type DbConfig struct {
	Type   string       `toml:"type"`
	MySql  MySqlConfig  `toml:"mysql"`
	Sqlite SqliteConfig `toml:"sqlite"`
}

type SqliteConfig struct {
	Debug bool `toml:"debug"`
}

type MySqlConfig struct {
	ConnectionString string        `toml:"connectionString"`
	MaxOpenConn      int           `toml:"maxOpenConn"`
	MaxIdleConn      int           `toml:"maxIdleConn"`
	ConnMaxLifeTime  time.Duration `toml:"connMaxLifeTime"`
	Debug            bool          `toml:"debug"`
}

type JWTConfig struct {
	AuthURL string `toml:"authURL"`
}

const (
	MinWaitingChainHeadStableDuration = time.Second * 2
	MaxWaitingChainHeadStableDuration = time.Second * 25
	DefWaitingChainHeadStableDuration = time.Second * 8
)

type MessageServiceConfig struct {
	WaitingChainHeadStableDuration time.Duration `toml:"WaitingChainHeadStableDuration"`

	SkipProcessHead bool `toml:"skipProcessHead"`
	SkipPushMessage bool `toml:"skipPushMessage"`
}

type Libp2pNetConfig struct {
	ListenAddress      string   `toml:"listenAddresses"`
	BootstrapAddresses []string `toml:"bootstrapAddresses"`
	// TODO: EnableRelay
	Enable bool `toml:"enablePubsub"`
}

type MessageStateConfig struct {
	BackTime int `toml:"backTime"` // 向前找多久的数据写到内存,单位秒

	DefaultExpiration, CleanupInterval int // message 缓存的有效时间和清理间隔
}

type GatewayConfig struct {
	Token string   `toml:"token"`
	Url   []string `toml:"url"`
}

type RateLimitConfig struct {
	Redis string `toml:"redis"`
}

func DefaultConfig() *Config {
	return &Config{
		DB: DbConfig{
			Type: "sqlite",
			MySql: MySqlConfig{
				ConnectionString: "",
				MaxOpenConn:      10,
				MaxIdleConn:      10,
				ConnMaxLifeTime:  time.Second * 60,
				Debug:            false,
			},
		},
		JWT: JWTConfig{
			AuthURL: "http://127.0.0.1:8989",
		},
		Log: LogConfig{
			Path:  "messager.log",
			Level: "info",
		},
		API: APIConfig{
			Address: "/ip4/0.0.0.0/tcp/39812",
		},
		Node: NodeConfig{
			Url:   "/ip4/127.0.0.1/tcp/3453",
			Token: "",
		},
		MessageService: MessageServiceConfig{
			WaitingChainHeadStableDuration: DefWaitingChainHeadStableDuration,

			SkipProcessHead: false,
			SkipPushMessage: false,
		},
		Gateway: GatewayConfig{
			Token: "",
			Url:   []string{"/ip4/127.0.0.1/tcp/45132"},
		},
		RateLimit: RateLimitConfig{Redis: ""},
		Trace:     metrics.DefaultTraceConfig(),
		Metrics:   metrics.DefaultMetricsConfig(),
		Libp2pNetConfig: &Libp2pNetConfig{
			ListenAddress:      "/ip4/0.0.0.0/tcp/0",
			BootstrapAddresses: []string{},
			Enable:             true,
		},
	}
}
