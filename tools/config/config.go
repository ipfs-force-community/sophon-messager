package config

type Config struct {
	Venus        VenusConfig
	Messager     MessagerConfig
	BatchReplace BatchReplaceConfig
}

type MessagerConfig struct {
	ConnectConfig
}

type VenusConfig struct {
	ConnectConfig
}

type ConnectConfig struct {
	URL   string
	Token string
}

type BatchReplaceConfig struct {
	BlockTime string
	Filters   Filters
	Methods   []int
}

type Filters struct {
	ActorCode string
	From      string
}

func DefaultConfig() *Config {
	return &Config{
		Messager: MessagerConfig{
			ConnectConfig: ConnectConfig{
				URL:   "/ip4/127.0.0.1/tcp/39812",
				Token: "",
			},
		},
		Venus: VenusConfig{
			ConnectConfig: ConnectConfig{
				URL:   "/ip4/127.0.0.1/tcp/3453",
				Token: "",
			},
		},
		BatchReplace: BatchReplaceConfig{
			BlockTime: "5m",
			Filters: Filters{
				ActorCode: "",
				From:      "",
			},
			Methods: []int{5},
		},
	}
}
