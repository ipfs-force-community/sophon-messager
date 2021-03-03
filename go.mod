module github.com/ipfs-force-community/venus-messager

go 1.15

require (
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/venus v0.9.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/hraban/lrucache v0.0.0-20201130153820-17052bf09781 // indirect
	github.com/hunjixin/automapper v0.0.0-20191127090318-9b979ce72ce2
	github.com/ipfs/go-cid v0.0.7
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/pelletier/go-toml v1.6.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/toorop/gin-logrus v0.0.0-20210225092905-2c785434f26f
	github.com/ugorji/go v1.2.4 // indirect
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/fx v1.13.1
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.12
	honnef.co/go/tools v0.1.2 // indirect
)

replace github.com/ipfs-force-community/venus-messager => ./

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
