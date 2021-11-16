package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
)

var oneTsStr = `{
    "Blocks": [{
        "miner": "t03112",
        "ticket": {
            "VRFProof": "hhugAMtLoAU1c0silWscHh6paIB2AmznnwPV3n+Y26SArhxasT7EcSIew98rTdJZC+bit6v4KkzwoGBsyJ8qNvxUxgcmMpL1/tzRZZFZ/CsH2IroIfbeNzjaNOUHJDin"
        },
        "electionProof": {
            "WinCount": 1,
            "VRFProof": "kZmZ80tYMFA+2BwDbTx6r5pVBrLP71JjAKV9odfr93nB222EzGvm8AsazqvsYPvkDaxhtFRQ1Kd2OyZ2A0rukMcHurQe0p09KeLYyBQchwXwQ+zch/Wx5tSDSBMBIz78"
        },
        "beaconEntries": [{
            "Round": 1239812,
            "Data": "i1vyxliz9NQdElFVnB65H9x9OKLE2sGTaDi25/7cxxg25oKt2BoE1pZMBLDUq8T+EIqsSdLmXKpQhw9esdOi4x5LbtcuzpZfv19ol1tQvBYbo0DGLvvJcemSCNcwYN5I"
        }],
        "winPoStProof": [{
            "PoStProof": 3,
            "ProofBytes": "iBCxcPGO2ZZK5FLXHAzB3/RWky1IRtUmPdG0YqZEa9OxLw8zqOogm5/x3wFkmp6qqBAssPxVQWAjQV6ov48r/3AZBCUAHfMHWaLyRdgDamBBTIcyE4MHmpTkTgOBG+JSBM8stSHABCbQE/TzzuzhyIYVPHH1VRVdhNevoXBuMF0ZWPWVamF7PO7+sNQfLzX/spUXFsl/hY0ZFxBfoV066wE8u8kTfHR1HqbfUXB6pmK/jrkyTG1ib3uL3kd9S2vp"
        }],
        "parents": [{
                "/": "bafy2bzacec75rx6r54nfeesdtqcnya5f4s6fp4i5ztwjvudacipvtc7x4blcu"
            },
            {
                "/": "bafy2bzaceawb4hfckvcyn6k5xb2pyqrl3qvt5ygm7nzobtzynh27nfa7eteos"
            },
            {
                "/": "bafy2bzaceczmi2d6gxdeotkdyfwq6zqyhp6jc5wkrvg32otaw3xxuso4o4ho6"
            },
            {
                "/": "bafy2bzaceaexstvj7t7l4uebe5w7d2xrqqrmkauh3dle4e6pknc2sz6omnx5q"
            },
            {
                "/": "bafy2bzacea5jvlzvkvf4ylyzc6bks3xiqjjuvnjftqnarlxxnlqczzpqzgry2"
            },
            {
                "/": "bafy2bzaceb3su3nciuoc65jw2fgl5x75u6cd3afoa2qsdjq7l2jm5tqagnoos"
            }
        ],
        "parentWeight": "5320219445",
        "height": 285488,
        "parentStateRoot": {
            "/": "bafy2bzaceahlyyvzs7sim6ss3lkoaregar3xwu4r5bousftlr7lqjokiiuagq"
        },
        "parentMessageReceipts": {
            "/": "bafy2bzacedswlcz5ddgqnyo3sak3jmhmkxashisnlpq6ujgyhe4mlobzpnhs6"
        },
        "messages": {
            "/": "bafy2bzacedmfo5tv5bmrsvinexz3yzwzrygh4tjk32zawvoioqif2qktzudek"
        },
        "BLSAggregate": {
            "Type": 2,
            "Data": "tFC8UCDEubhyqYmGoE9wz93Ntd8obhVvgQbWe8cq/APYJHWf9799jvZCyNVvTuptFfSmuxOOZgm89uPkkMVDSfOfF3EZTxnL/9JLG+CsuJA/GZ5nb8Phjmd5zM6AzXtf"
        },
        "timestamp": 1632625440,
        "blocksig": {
            "Type": 2,
            "Data": "sXqkz37X05Togm424inkIRmRacgOnxJD1wDC1VWTkSu1eNrRkDKqnB22CDlEX2GuGbs2k7t+HVpp/JKkFLprUf10KFj/O5v4zbefadyAoJZlEB5BOL48NiXhK9YXDg4Z"
        },
        "forkSignaling": 0,
        "parentBaseFee": "100"
    }],
    "Key": [{
        "/": "bafy2bzacec5m52bayhakga37ndi7kygqgfqezwkn5njkpam3prsxnrpsq543s"
    }]
}`

var TwoTsStr = `{
    "Blocks": [{
        "miner": "t02000",
        "ticket": {
            "VRFProof": "r6MAcSUoqdxN9nOuuhGyFaUnGE6hAT0E3Q4RfUtnXG3NpLNqptfG0TN6qsPNozi1BMDYpChptyBzvMDL5igKrH7rVdMtqk9hhwb4AuW+yQPK4Bf92Q5Nq73fLt6qF4iM"
        },
        "electionProof": {
            "WinCount": 1,
            "VRFProof": "pIyRJSG2ngB+L5vPkvWLc6cGA5XwqZna5SoPybMwxfkaFPGHWN3C/5DcjHOqVdWBA9li80+kK1U5oJPH71jey8xnGG0CoCcwYbfOPv7iw6YAhMl2jUyMx633zHqJ3Zwe"
        },
        "beaconEntries": [{
            "Round": 1239813,
            "Data": "gsAJiXZ+PTkCAV7PkfBWNIkAST1b5sq5Khia8K8Pt5oUwDYGCnS/p0yyf+jHs27YEkpK1KQOnZLRQhOcCOkpvzGVs1s7xK2sxfbK17XWBeQm0m3CJAEGnWyV4Jvk/IV4"
        }],
        "winPoStProof": [{
            "PoStProof": 3,
            "ProofBytes": "tJe/W0PYAT89ztMf01VaQUehrqxcqNPhEcYTD24QVBT6fEOsR19ET0jen0XXZEJGgdDMKprJTSozRFBfNKhe9T5s921wtleaUiDY/PP3jM8kGNE8UFDzWAfRp6xTCgadAdtBkUgppjAFJGEIY16YzoWtn6b6ji9Xyx/NlP463+KCd2GgnQ+p/LQwEpVzk736qMagLcT9Sxf9Nx5A5jxvt6HR64RBTXKXN9efgvZONYWgriocMrT0hGIvl2mqGcYi"
        }],
        "parents": [{
            "/": "bafy2bzacec5m52bayhakga37ndi7kygqgfqezwkn5njkpam3prsxnrpsq543s"
        }],
        "parentWeight": "5320234088",
        "height": 285489,
        "parentStateRoot": {
            "/": "bafy2bzaceagbmtfuyqg53vib42cvhxfzb5izyy47b3gvjwhnwtpwfdccjowxu"
        },
        "parentMessageReceipts": {
            "/": "bafy2bzacecya3c3rswlgasmdwr3jg4k6hvxkqd6mmrlvayswws6c22syprctu"
        },
        "messages": {
            "/": "bafy2bzacedja7rjuwkpzlkvuznhehdo3s4ajwkqgv5xyzo3rfbswfzv4pugbi"
        },
        "BLSAggregate": {
            "Type": 2,
            "Data": "rNJrszC8RhY+udWuXUSPUPtD3wvKe3/mS6QsgwBrpyRNlwDjnV5k47Y2phzldKXQGZ1+2H0VeEeQ0iwrUCaoStBSXWXQjjfQxkt9CAPrAquqI2/fzc1dDHT3bogJ+1CV"
        },
        "timestamp": 1632625470,
        "blocksig": {
            "Type": 2,
            "Data": "tfeAFzp87TNl2gVPjmY3i6OZ2CeTf1v4YYj7Vtf+f7j/m5HfO4/NwqkdOfrlYT+ID8gdCFHlj/94JmICGaByHFFk4HejL+x1OrR2IB3uB2Aio6roS2nBtuV1KE8QmXEx"
        },
        "forkSignaling": 0,
        "parentBaseFee": "100"
    }],
    "Key": [{
        "/": "bafy2bzaceabcn7hde6vvzqgmwowioew3ycv3l2u2onwkemw673aribydelca4"
    }]
}`

func TestReadAndWriteTipset(t *testing.T) {
	tsCache := &TipsetCache{Cache: map[int64]*venusTypes.TipSet{}, CurrHeight: 0}

	ts := &venusTypes.TipSet{}
	assert.Nil(t, ts.UnmarshalJSON([]byte(oneTsStr)))
	ts2 := &venusTypes.TipSet{}
	assert.Nil(t, ts2.UnmarshalJSON([]byte(TwoTsStr)))
	tsCache.Add(ts, ts2)

	filePath := "./test_read_write_tipset.json"
	defer func() {
		assert.NoError(t, os.Remove(filePath))
	}()
	err := tsCache.Save(filePath)
	assert.NoError(t, err)

	cache2 := newTipsetCache()
	err = cache2.Load(filePath)
	assert.NoError(t, err)
	assert.Equal(t, tsCache, cache2)
}
