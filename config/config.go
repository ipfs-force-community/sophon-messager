package config

import "time"

type Config struct {
	DB             DbConfig             `toml:"db"`
	JWT            JWTConfig            `toml:"jwt"`
	Log            LogConfig            `toml:"log"`
	API            APIConfig            `toml:"api"`
	Node           NodeConfig           `toml:"node"`
	MessageService MessageServiceConfig `toml:messageService`
	MessageState   MessageStateConfig   `toml:messageState`
	Address        AddressConfig        `toml:"address"`
}

type NodeConfig struct {
	Url   string `toml:"url"`
	Token string `toml:"token"`
}

type LogConfig struct {
	Path string `toml:"path"`
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
	Path string `toml:"path"`
}

type MySqlConfig struct {
	Addr            string        `toml:"addr"`
	User            string        `toml:"user"`
	Pass            string        `toml:"pass"`
	Name            string        `toml:"name"`
	MaxOpenConn     int           `toml:"maxOpenConn"`
	MaxIdleConn     int           `toml:"maxIdleConn"`
	ConnMaxLifeTime time.Duration `toml:"connMaxLifeTime"`
}

type JWTConfig struct {
	Secret         string        `toml:"secret"`
	ExpireDuration time.Duration `toml:"expireDuration"`
}

type AddressConfig struct {
	RemoteWalletSweepInterval time.Duration `toml:"remoteWalletSweepInterval"` // second
}

type MessageServiceConfig struct {
	TipsetFilePath string `toml:"tipsetFilePath"`
}

type MessageStateConfig struct {
	BackTime time.Duration `toml:"backTime"` // 向前找多久的数据写到内存,单位秒

	DefaultExpiration, CleanupInterval time.Duration // message 缓存的有效时间和清理间隔
}
