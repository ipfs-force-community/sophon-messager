package config

import (
	"time"

	"github.com/ipfs-force-community/metrics"

	gatewayTypes "github.com/ipfs-force-community/venus-gateway/types"
)

type Config struct {
	DB             DbConfig             `toml:"db"`
	JWT            JWTConfig            `toml:"jwt"`
	Log            LogConfig            `toml:"log"`
	API            APIConfig            `toml:"api"`
	Node           NodeConfig           `toml:"node"`
	MessageService MessageServiceConfig `toml:"messageService"`
	MessageState   MessageStateConfig   `toml:"messageState"`
	Gateway        GatewayConfig        `toml:"gateway"`
	RateLimit      RateLimitConfig      `toml:"rateLimit"`
	Trace          metrics.TraceConfig  `toml:"tracing"`
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
	File  string `toml:"file"`
	Debug bool   `toml:"debug"`
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
	Local   struct {
		Secret string `toml:"secret"`
		Token  string `toml:"token"`
	} `toml:"local"`
}

const MinWaitingChainHeadStableDuration,
MaxWaitingChainHeadStableDuration,
DefWaitingChainHeadStableDuration = time.Second * 2, time.Second * 25, time.Second * 8

type MessageServiceConfig struct {
	WaitingChainHeadStableDuration time.Duration `toml:"WaitingChainHeadStableDuration"`

	TipsetFilePath  string `toml:"tipsetFilePath"`
	SkipProcessHead bool   `toml:"skipProcessHead"`
	SkipPushMessage bool   `toml:"skipPushMessage"`
}

type MessageStateConfig struct {
	BackTime int `toml:"backTime"` // 向前找多久的数据写到内存,单位秒

	DefaultExpiration, CleanupInterval int // message 缓存的有效时间和清理间隔
}

type GatewayConfig struct {
	RemoteEnable bool                `toml:"remoteEnable"`
	Token        string              `toml:"token"`
	Url          []string              `toml:"url"`
	Cfg          gatewayTypes.Config `toml:"cfg"`
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
			Sqlite: SqliteConfig{File: "./message.db"},
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
		MessageState: MessageStateConfig{
			BackTime:          3600 * 24,
			DefaultExpiration: 3600 * 24 * 3,
			CleanupInterval:   3600 * 24,
		},
		MessageService: MessageServiceConfig{
			WaitingChainHeadStableDuration: DefWaitingChainHeadStableDuration,

			TipsetFilePath:  "./tipset.json",
			SkipProcessHead: false,
			SkipPushMessage: false,
		},
		Gateway: GatewayConfig{
			RemoteEnable: true,
			Token:        "",
			Url:          []string{"/ip4/127.0.0.1/tcp/45132"},
			Cfg: gatewayTypes.Config{
				RequestQueueSize: 30,
				RequestTimeout:   time.Minute * 5,
			},
		},
		RateLimit: RateLimitConfig{Redis: ""},
		Trace: metrics.TraceConfig{
			JaegerEndpoint:       "",
			ProbabilitySampler:   1.0,
			JaegerTracingEnabled: false,
			ServerName:           "venus-messenger",
		},
	}
}
