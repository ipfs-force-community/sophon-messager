module github.com/filecoin-project/venus-messager

go 1.15

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/fatih/color v1.10.0
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/venus v1.0.5-0.20211011075109-4c9cb88878f3
	github.com/filecoin-project/venus-auth v1.3.1-0.20210809053831-012d55d5f578
	github.com/filecoin-project/venus-wallet v1.2.2-0.20211011030242-5037d6297fa3
	github.com/gbrlsnchs/jwt/v3 v3.0.0
	github.com/google/uuid v1.2.0
	github.com/hraban/lrucache v0.0.0-20201130153820-17052bf09781 // indirect
	github.com/hunjixin/automapper v0.0.0-20191127090318-9b979ce72ce2
	github.com/ipfs-force-community/metrics v1.0.1-0.20210827074542-cc8db7683f13
	github.com/ipfs-force-community/venus-common-utils v0.0.0-20210714054928-2042a9040759
	github.com/ipfs-force-community/venus-gateway v1.1.2-0.20210731031356-770f19abfbcb
	github.com/ipfs/go-cid v0.0.7
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.6.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210219115102-f37d292932f2
	go.uber.org/fx v1.13.1
	golang.org/x/exp v0.0.0-20200513190911-00229845015e // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gorm.io/driver/mysql v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.12
	modernc.org/mathutil v1.1.1
)

replace github.com/filecoin-project/go-jsonrpc => github.com/ipfs-force-community/go-jsonrpc v0.1.4-0.20210731021807-68e5207079bc

replace github.com/ipfs/go-ipfs-cmds => github.com/ipfs-force-community/go-ipfs-cmds v0.6.1-0.20210521090123-4587df7fa0ab
