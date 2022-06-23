module github.com/filecoin-project/venus-messager

go 1.15

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-address v0.0.6
	github.com/filecoin-project/go-cbor-util v0.0.1
	github.com/filecoin-project/go-jsonrpc v0.1.5
	github.com/filecoin-project/go-state-types v0.1.10
	github.com/filecoin-project/specs-actors/v6 v6.0.2
	github.com/filecoin-project/venus v1.6.0-rc2
	github.com/filecoin-project/venus-auth v1.6.0-rc1
	github.com/gbrlsnchs/jwt/v3 v3.0.1
	github.com/google/uuid v1.3.0
	github.com/hraban/lrucache v0.0.0-20201130153820-17052bf09781 // indirect
	github.com/hunjixin/automapper v0.0.0-20191127090318-9b979ce72ce2
	github.com/ipfs-force-community/metrics v1.0.1-0.20211228055608-9462dc86e157
	github.com/ipfs-force-community/venus-common-utils v0.0.0-20210924063144-1d3a5b30de87
	github.com/ipfs-force-community/venus-gateway v1.6.0-rc2
	github.com/ipfs/go-cid v0.1.0
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.5.0
	github.com/onsi/gomega v1.16.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.9.4
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.9.0 // indirect
	github.com/stretchr/testify v1.7.1
	github.com/ugorji/go v1.2.4 // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20220323183124-98fa8256a799
	go.uber.org/fx v1.15.0
	gorm.io/driver/mysql v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.12
	modernc.org/mathutil v1.1.1
)

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20211201033628-fc1430d095f6
