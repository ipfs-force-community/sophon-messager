package utils

import (
	"strings"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
)

func StringToTipsetKey(str string) (venusTypes.TipSetKey, error) {
	str = strings.TrimLeft(str, "{ ")
	str = strings.TrimRight(str, " }")
	var cids []cid.Cid
	for _, s := range strings.Split(str, " ") {
		c, err := cid.Decode(s)
		if err != nil {
			return venusTypes.TipSetKey{}, err
		}
		cids = append(cids, c)
	}

	return venusTypes.NewTipSetKey(cids...), nil
}
