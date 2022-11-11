package metrics

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Global Tags
var (
	WalletAddress, _ = tag.NewKey("wallet")
)

// Distribution
var defaultSecondsDistribution = view.Distribution(8, 9, 10, 12, 14, 16, 18, 20, 25, 30, 60)

var (
	WalletBalance    = stats.Float64("wallet_balance", "Wallet balance", stats.UnitDimensionless)
	WalletDBNonce    = stats.Int64("wallet_db_nonce", "Wallet nonce in db", stats.UnitDimensionless)
	WalletChainNonce = stats.Int64("wallet_chain_nonce", "Wallet nonce on the chain", stats.UnitDimensionless)

	NumOfUnFillMsg = stats.Int64("num_of_unfill_msg", "The number of unFill msg", stats.UnitDimensionless)
	NumOfFillMsg   = stats.Int64("num_of_fill_msg", "The number of fill Msg", stats.UnitDimensionless)
	NumOfFailedMsg = stats.Int64("num_of_failed_msg", "The number of failed msg", stats.UnitDimensionless)

	NumOfMsgBlockedThreeMinutes = stats.Int64("blocked_three_minutes_msgs", "Number of messages blocked for more than 3 minutes", stats.UnitDimensionless)
	NumOfMsgBlockedFiveMinutes  = stats.Int64("blocked_five_minutes_msgs", "Number of messages blocked for more than 5 minutes", stats.UnitDimensionless)

	SelectedMsgNumOfLastRound = stats.Int64("selected_msg_num", "Number of selected messages in the last round", stats.UnitDimensionless)
	ToPushMsgNumOfLastRound   = stats.Int64("topush_msg_num", "Number of to-push messages in the last round", stats.UnitDimensionless)
	ErrMsgNumOfLastRound      = stats.Int64("err_msg_num", "Number of err messages in the last round", stats.UnitDimensionless)

	ChainHeadStableDelay    = stats.Int64("chain_head_stable_s", "Delay of chain head stabilization", stats.UnitSeconds)
	ChainHeadStableDuration = stats.Int64("chain_head_stable_dur_s", "Duration of chain head stabilization", stats.UnitSeconds)
)

var (
	WalletBalanceView = &view.View{
		Measure:     WalletBalance,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	WalletChainNonceView = &view.View{
		Measure:     WalletChainNonce,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	WalletDBNonceView = &view.View{
		Measure:     WalletDBNonce,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}

	NumOfUnFillMsgView = &view.View{
		Measure:     NumOfUnFillMsg,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	NumOfFillMsgView = &view.View{
		Measure:     NumOfFillMsg,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	NumOfFailedMsgView = &view.View{
		Measure:     NumOfFailedMsg,
		Aggregation: view.LastValue(),
	}

	NumOfMsgBlockedThreeMinutesView = &view.View{
		Measure:     NumOfMsgBlockedThreeMinutes,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	NumOfMsgBlockedFiveMinutesView = &view.View{
		Measure:     NumOfMsgBlockedFiveMinutes,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}

	SelectedMsgNumOfLastRoundView = &view.View{
		Measure:     SelectedMsgNumOfLastRound,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	ToPushMsgNumOfLastRoundView = &view.View{
		Measure:     ToPushMsgNumOfLastRound,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}
	ErrMsgNumOfLastRoundView = &view.View{
		Measure:     ErrMsgNumOfLastRound,
		Aggregation: view.LastValue(),
		TagKeys:     []tag.Key{WalletAddress},
	}

	ChainHeadStableDelayView = &view.View{
		Measure:     ChainHeadStableDelay,
		Aggregation: view.LastValue(),
	}
	ChainHeadStableDurationView = &view.View{
		Measure:     ChainHeadStableDuration,
		Aggregation: defaultSecondsDistribution,
	}
)

var MessagerNodeViews = []*view.View{
	WalletBalanceView,
	WalletChainNonceView,
	WalletDBNonceView,

	NumOfUnFillMsgView,
	NumOfFillMsgView,
	NumOfFailedMsgView,

	NumOfMsgBlockedThreeMinutesView,
	NumOfMsgBlockedFiveMinutesView,

	SelectedMsgNumOfLastRoundView,
	ToPushMsgNumOfLastRoundView,
	ErrMsgNumOfLastRoundView,

	ChainHeadStableDelayView,
	ChainHeadStableDurationView,
}
