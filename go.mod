module github.com/filecoin-project/venus-messager

go 1.15

require (
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-cbor-util v0.0.0-20201016124514-d0bbec7bfcc4
	github.com/filecoin-project/go-fil-commcid v0.1.0 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/specs-actors v0.9.14
	github.com/filecoin-project/specs-actors/v2 v2.3.5
	github.com/filecoin-project/specs-actors/v3 v3.1.1
	github.com/filecoin-project/specs-actors/v4 v4.0.1
	github.com/filecoin-project/specs-actors/v5 v5.0.4
	github.com/filecoin-project/specs-actors/v6 v6.0.0
	github.com/filecoin-project/venus v1.1.2-rc2
	github.com/filecoin-project/venus-auth v1.3.1
	github.com/gbrlsnchs/jwt/v3 v3.0.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hraban/lrucache v0.0.0-20201130153820-17052bf09781 // indirect
	github.com/hunjixin/automapper v0.0.0-20191127090318-9b979ce72ce2
	github.com/ipfs-force-community/metrics v1.0.1-0.20211022060227-11142a08b729
	github.com/ipfs-force-community/venus-common-utils v0.0.0-20210714054928-2042a9040759
	github.com/ipfs-force-community/venus-gateway v1.1.2-0.20210731031356-770f19abfbcb
	github.com/ipfs/go-blockservice v0.1.5 // indirect
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-ipfs-blockstore v1.0.4 // indirect
	github.com/ipld/go-car v0.3.2-0.20211001225732-32d0d9933823 // indirect
	github.com/ipld/go-ipld-prime v0.12.3 // indirect
	github.com/klauspost/cpuid/v2 v2.0.8 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/onsi/gomega v1.16.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.9.4
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.9.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	go.uber.org/fx v1.13.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.5.0 // indirect
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gorm.io/driver/mysql v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.12
	modernc.org/mathutil v1.1.1
)

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20211201033628-fc1430d095f6

replace github.com/ipfs/go-ipfs-cmds => github.com/ipfs-force-community/go-ipfs-cmds v0.6.1-0.20210521090123-4587df7fa0ab
