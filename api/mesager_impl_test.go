package api

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc/auth"
	vauth "github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus-messager/mocks"
	"github.com/filecoin-project/venus-messager/testhelper"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/golang/mock/gomock"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
)

func TestListMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListMessage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams) {
			dict := param.ToMap()
			assert.Equal(t, 0, len(dict))
		})
		_, err := p.impl.ListMessage(p.ctxAdmin, &types.MsgQueryParams{})
		assert.NoError(t, err)

	})

	t.Run("normal user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListMessage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams) {
			dict := param.ToMap()
			assert.Equal(t, 1, len(dict))
			assert.Equal(t, p.userR, dict["wallet_name"])
		})
		_, err := p.impl.ListMessage(p.ctxUserR, &types.MsgQueryParams{})
		assert.NoError(t, err)
	})

	t.Run("no user", func(t *testing.T) {
		_, err := p.impl.ListMessage(p.ctx, &types.MsgQueryParams{})
		assert.Equal(t, ErrorUserNotFound, err)
	})
}

func TestListFailedMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListFailedMessage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams) {
			dict := param.ToMap()
			assert.Equal(t, 0, len(dict))
		})
		_, err := p.impl.ListFailedMessage(p.ctxAdmin)
		assert.NoError(t, err)

	})

	t.Run("normal user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListFailedMessage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams) {
			dict := param.ToMap()
			assert.Equal(t, 1, len(dict))
			assert.Equal(t, p.userR, dict["wallet_name"])
		})
		_, err := p.impl.ListFailedMessage(p.ctxUserR)
		assert.NoError(t, err)
	})

	t.Run("no user", func(t *testing.T) {
		_, err := p.impl.ListFailedMessage(p.ctx)
		assert.Equal(t, ErrorUserNotFound, err)
	})
}

func TestListBlockedMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListBlockedMessage(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams, d time.Duration) {
			dict := param.ToMap()
			assert.Equal(t, 1, len(dict))
		})

		_, err := p.impl.ListBlockedMessage(p.ctxAdmin, p.addr1, time.Second)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().ListBlockedMessage(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, param *types.MsgQueryParams, d time.Duration) {
			dict := param.ToMap()
			assert.Equal(t, 1, len(dict))
		})

		_, err := p.impl.ListBlockedMessage(p.ctxUserR, p.addr1, time.Second)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.ListBlockedMessage(p.ctxUserW, p.addr2, time.Second)
		assert.Equal(t, ErrorPermissionDeny, err)
	})

	t.Run("no user", func(t *testing.T) {
		_, err := p.impl.ListBlockedMessage(p.ctx, p.addr2, time.Second)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestGetMessageByUid(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		_, err := p.impl.GetMessageByUid(p.ctxAdmin, "message_id")
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		_, err := p.impl.GetMessageByUid(p.ctxUserR, "message_id")
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.GetMessageByUid(p.ctxUserW, "message_id")
		assert.Equal(t, ErrorPermissionDeny, err)
	})

	t.Run("no user", func(t *testing.T) {
		_, err := p.impl.GetMessageByUid(p.ctx, "message_id")
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestGetMessageBySignedCid(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageBySignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.SignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageBySignedCid(p.ctxAdmin, test_cid)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageBySignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.SignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageBySignedCid(p.ctxUserR, test_cid)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageBySignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.SignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageBySignedCid(p.ctxUserW, test_cid)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestGetMessageByUnsignedCid(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByUnsignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.UnsignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageByUnsignedCid(p.ctxAdmin, test_cid)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByUnsignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.UnsignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageByUnsignedCid(p.ctxUserR, test_cid)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByUnsignedCid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id cid.Cid) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.UnsignedCid = &id
			msg.WalletName = p.userR
			return msg, nil
		})
		test_cid := cid.NewCidV1(cid.Raw, []byte("test"))
		_, err := p.impl.GetMessageByUnsignedCid(p.ctxUserW, test_cid)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestGetMessageByFromAndNonce(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByFromAndNonce(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.WalletName = p.userR
			return msg, nil
		})
		_, err := p.impl.GetMessageByFromAndNonce(p.ctxAdmin, p.addr1, 1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByFromAndNonce(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.WalletName = p.userR
			return msg, nil
		})
		_, err := p.impl.GetMessageByFromAndNonce(p.ctxUserR, p.addr1, 1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		p.msgSrv.EXPECT().GetMessageByFromAndNonce(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
			msg := testhelper.NewMessage()
			msg.WalletName = p.userR
			return msg, nil
		})
		_, err := p.impl.GetMessageByFromAndNonce(p.ctxUserW, p.addr1, 1)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestRecoverFailedMsg(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().RecoverFailedMsg(gomock.Any(), gomock.Any())
		_, err := p.impl.RecoverFailedMsg(p.ctxAdmin, p.addr2)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().RecoverFailedMsg(gomock.Any(), gomock.Any())

		_, err := p.impl.RecoverFailedMsg(p.ctxUserR, p.addr2)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.RecoverFailedMsg(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestRepublishMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().RepublishMessage(gomock.Any(), gomock.Any())
		err := p.impl.RepublishMessage(p.ctxAdmin, "message_id")
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().RepublishMessage(gomock.Any(), gomock.Any())
		err := p.impl.RepublishMessage(p.ctxUserR, "message_id")
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.RepublishMessage(p.ctxUserW, "message_id")
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestUpdateMessageStateByID(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().UpdateMessageStateByID(gomock.Any(), gomock.Any(), gomock.Any())
		err := p.impl.UpdateMessageStateByID(p.ctxAdmin, "message_id", types.UnFillMsg)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().UpdateMessageStateByID(gomock.Any(), gomock.Any(), gomock.Any())
		err := p.impl.UpdateMessageStateByID(p.ctxUserR, "message_id", types.UnFillMsg)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.UpdateMessageStateByID(p.ctxUserW, "message_id", types.UnFillMsg)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestMarkBadMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().MarkBadMessage(gomock.Any(), gomock.Any())
		err := p.impl.MarkBadMessage(p.ctxAdmin, "message_id")
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().MarkBadMessage(gomock.Any(), gomock.Any())
		err := p.impl.MarkBadMessage(p.ctxUserR, "message_id")
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.MarkBadMessage(p.ctxUserW, "message_id")
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestUpdateFilledMessageByID(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.msgSrv.EXPECT().UpdateFilledMessageByID(gomock.Any(), gomock.Any())
		_, err := p.impl.UpdateFilledMessageByID(p.ctxAdmin, "message_id")
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().UpdateFilledMessageByID(gomock.Any(), gomock.Any())
		_, err := p.impl.UpdateFilledMessageByID(p.ctxUserR, "message_id")
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.UpdateFilledMessageByID(p.ctxUserW, "message_id")
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestWaitMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {

		p.msgSrv.EXPECT().WaitMessage(gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.WaitMessage(p.ctxAdmin, "message_id", 1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.msgSrv.EXPECT().WaitMessage(gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.WaitMessage(p.ctxUserR, "message_id", 1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.WaitMessage(p.ctxUserW, "message_id", 1)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestPushMessage(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		p.msgSrv.EXPECT().PushMessage(gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.PushMessage(p.ctxAdmin, &msg.Message, &types.SendSpec{})
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		p.msgSrv.EXPECT().PushMessage(gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.PushMessage(p.ctxUserR, &msg.Message, &types.SendSpec{})
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		_, err := p.impl.PushMessage(p.ctxUserW, &msg.Message, &types.SendSpec{})
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestPushMessageWithId(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		p.msgSrv.EXPECT().PushMessageWithId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.PushMessageWithId(p.ctxAdmin, "msg_id", &msg.Message, &types.SendSpec{})
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		p.msgSrv.EXPECT().PushMessageWithId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
		_, err := p.impl.PushMessageWithId(p.ctxUserR, "msg_id", &msg.Message, &types.SendSpec{})
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		msg := testhelper.NewMessage()
		msg.From = p.addr2
		_, err := p.impl.PushMessageWithId(p.ctxUserW, "msg_id", &msg.Message, &types.SendSpec{})
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestHasMessageByUid(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		ok, err := p.impl.HasMessageByUid(p.ctxAdmin, "message_id")
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		ok, err := p.impl.HasMessageByUid(p.ctxUserR, "message_id")
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		ok, err := p.impl.HasMessageByUid(p.ctxUserW, "message_id")
		assert.False(t, ok)
		assert.Error(t, err)
	})
}

func TestGetAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().GetAddress(gomock.Any(), gomock.Any())
		_, err := p.impl.GetAddress(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().GetAddress(gomock.Any(), gomock.Any())
		_, err := p.impl.GetAddress(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.GetAddress(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestListAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().ListAddress(gomock.Any()).DoAndReturn(func(ctx context.Context) ([]*types.Address, error) {
			return []*types.Address{
				{
					Addr: p.addr1,
				},
				{
					Addr: p.addr2,
				},
			}, nil
		})

		addrs, err := p.impl.ListAddress(p.ctxAdmin)
		assert.Len(t, addrs, 2)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().ListAddress(gomock.Any()).DoAndReturn(func(ctx context.Context) ([]*types.Address, error) {
			return []*types.Address{
				{
					Addr: p.addr1,
				},
				{
					Addr: p.addr2,
				},
			}, nil
		})

		addrs, err := p.impl.ListAddress(p.ctxUserR)
		assert.Len(t, addrs, 2)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		p.addrSrv.EXPECT().ListAddress(gomock.Any()).DoAndReturn(func(ctx context.Context) ([]*types.Address, error) {
			return []*types.Address{
				{
					Addr: p.addr1,
				},
				{
					Addr: p.addr2,
				},
			}, nil
		})

		addrs, err := p.impl.ListAddress(p.ctxUserW)
		assert.Len(t, addrs, 1)
		assert.NoError(t, err)
	})
}

func TestUpdateNonce(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().UpdateNonce(gomock.Any(), gomock.Any(), gomock.Any())
		err := p.impl.UpdateNonce(p.ctxAdmin, p.addr1, 1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().UpdateNonce(gomock.Any(), gomock.Any(), gomock.Any())
		err := p.impl.UpdateNonce(p.ctxUserR, p.addr1, 1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.UpdateNonce(p.ctxUserW, p.addr2, 1)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestDeleteAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().DeleteAddress(gomock.Any(), gomock.Any())
		err := p.impl.DeleteAddress(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().DeleteAddress(gomock.Any(), gomock.Any())
		err := p.impl.DeleteAddress(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.DeleteAddress(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestForbiddenAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().ForbiddenAddress(gomock.Any(), gomock.Any())
		err := p.impl.ForbiddenAddress(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().ForbiddenAddress(gomock.Any(), gomock.Any())
		err := p.impl.ForbiddenAddress(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.ForbiddenAddress(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestActiveAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().ActiveAddress(gomock.Any(), gomock.Any())
		err := p.impl.ActiveAddress(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().ActiveAddress(gomock.Any(), gomock.Any())
		err := p.impl.ActiveAddress(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.ActiveAddress(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestSetFeeParams(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().SetFeeParams(gomock.Any(), gomock.Any())
		err := p.impl.SetFeeParams(p.ctxAdmin, &types.AddressSpec{Address: p.addr2})
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().SetFeeParams(gomock.Any(), gomock.Any())
		err := p.impl.SetFeeParams(p.ctxUserR, &types.AddressSpec{Address: p.addr2})
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		err := p.impl.SetFeeParams(p.ctxUserW, &types.AddressSpec{Address: p.addr2})
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestHasAddress(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().HasAddress(gomock.Any(), gomock.Any())
		_, err := p.impl.HasAddress(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().HasAddress(gomock.Any(), gomock.Any())
		_, err := p.impl.HasAddress(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.HasAddress(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

func TestWalletHas(t *testing.T) {
	p := getTestParams(t)

	t.Run("admin user", func(t *testing.T) {
		p.addrSrv.EXPECT().WalletHas(gomock.Any(), gomock.Any())
		_, err := p.impl.WalletHas(p.ctxAdmin, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("right user", func(t *testing.T) {
		p.addrSrv.EXPECT().WalletHas(gomock.Any(), gomock.Any())
		_, err := p.impl.WalletHas(p.ctxUserR, p.addr1)
		assert.NoError(t, err)
	})

	t.Run("wrong user", func(t *testing.T) {
		_, err := p.impl.WalletHas(p.ctxUserW, p.addr2)
		assert.Equal(t, ErrorPermissionDeny, err)
	})
}

type msgSrvTestParams struct {
	ctx        context.Context
	impl       *MessageImp
	addrSrv    *mocks.MockIAddressService
	msgSrv     *mocks.MockIMessageService
	authClient *mocks.MockIAuthClient
	userR      string
	userW      string
	ctxUserR   context.Context
	ctxUserW   context.Context
	ctxAdmin   context.Context
	addr1      address.Address
	addr2      address.Address
	addr2user  map[string][]string
}

func getTestParams(t *testing.T) *msgSrvTestParams {

	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	addrSrv := mocks.NewMockIAddressService(ctrl)
	msgSrv := mocks.NewMockIMessageService(ctrl)
	authClient := mocks.NewMockIAuthClient(ctrl)

	msgImpl := MessageImp{
		AddressSrv: addrSrv,
		MessageSrv: msgSrv,
		AuthClient: authClient,
	}

	userR := randString(10)
	userW := randString(10)
	addr1 := testhelper.RandAddresses(t, 1)[0]
	addr2 := testhelper.RandAddresses(t, 1)[0]
	addr2user := map[string][]string{
		addr1.String(): {userR, userW},
		addr2.String(): {
			userR,
		},
	}

	ctxUserR := jwtclient.CtxWithName(ctx, userR)
	ctxUserW := jwtclient.CtxWithName(ctx, userW)
	ctxAdmin := auth.WithPerm(ctx, []string{"admin", "write", "read"})

	msgSrv.EXPECT().GetMessageByUid(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id string) (*types.Message, error) {
		msg := testhelper.NewMessage()
		msg.ID = id
		msg.WalletName = userR
		return msg, nil
	}).AnyTimes()

	authClient.EXPECT().GetUserBySigner(gomock.Any()).DoAndReturn(func(addr string) ([]*vauth.OutputUser, error) {
		ret := []*vauth.OutputUser{}
		for _, user := range addr2user[addr] {
			ret = append(ret, &vauth.OutputUser{
				Name: user,
			})
		}
		return ret, nil
	}).AnyTimes()

	return &msgSrvTestParams{
		ctx:        ctx,
		impl:       &msgImpl,
		addrSrv:    addrSrv,
		msgSrv:     msgSrv,
		authClient: authClient,
		userR:      userR,
		userW:      userW,
		ctxUserR:   ctxUserR,
		ctxUserW:   ctxUserW,
		ctxAdmin:   ctxAdmin,
		addr1:      addr1,
		addr2:      addr2,
		addr2user:  addr2user,
	}
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
