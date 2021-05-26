package config

import (
	"time"

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
	Wallet         WalletConfig         `toml:"wallet"`
	Gateway        GatewayConfig        `toml:"gateway"`
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
	Path  string `toml:"path"`
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
	Url string `toml:"url"`
}

type WalletConfig struct {
	ScanInterval int `toml:"scanInterval"` // second
}

type MessageServiceConfig struct {
	TipsetFilePath  string `toml:"tipsetFilePath"`
	SkipProcessHead bool   `toml:"skipProcessHead"`
	SkipPushMessage bool   `toml:"skipPushMessage"`
}

type MessageStateConfig struct {
	BackTime int `toml:"backTime"` // 向前找多久的数据写到内存,单位秒

	DefaultExpiration, CleanupInterval int // message 缓存的有效时间和清理间隔
}

type GatewayConfig struct {
	Disable bool                `toml:"disable"`
	Token   string              `toml:"token"`
	Url     string              `toml:"url"`
	Cfg     gatewayTypes.Config `toml:"cfg"`
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
			Sqlite: SqliteConfig{Path: "./message.db"},
		},
		JWT: JWTConfig{
			Url: "http://127.0.0.1:8989",
		},
		Log: LogConfig{
			Path:  "messager.log",
			Level: "info",
		},
		API: APIConfig{
			Address: "/ip4/0.0.0.0/tcp/39812",
		},
		Node: NodeConfig{
			Url:   "",
			Token: "",
		},
		Wallet: WalletConfig{
			ScanInterval: 10,
		},
		MessageState: MessageStateConfig{
			BackTime:          3600 * 24,
			DefaultExpiration: 3600 * 24 * 3,
			CleanupInterval:   3600 * 24,
		},
		MessageService: MessageServiceConfig{
			TipsetFilePath:  "./tipset.json",
			SkipProcessHead: false,
			SkipPushMessage: false,
		},
		Gateway: GatewayConfig{
			Disable: false,
			Token:   "",
			Url:     "",
			Cfg: gatewayTypes.Config{
				RequestQueueSize: 30,
				RequestTimeout:   time.Minute * 5,
			},
		},
	}
}
