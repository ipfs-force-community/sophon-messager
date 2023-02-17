package testhelper

import (
	"context"
	"errors"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient/mocks"
	"github.com/golang/mock/gomock"

	"github.com/filecoin-project/venus-auth/jwtclient"
)

type AuthClient struct {
	*mocks.MockIAuthClient
}

func (m *AuthClient) Init(account string, addrs []address.Address) {
	signers := make(map[address.Address]map[string]struct{})
	for _, signer := range addrs {
		if signer.Protocol() == address.ID {
			signer, _ = ResolveIDAddr(signer)
		}

		users, ok := signers[signer]
		if !ok {
			newUsers := make(map[string]struct{})
			newUsers[account] = struct{}{}
			signers[signer] = newUsers
		} else {
			if _, ok := users[account]; !ok {
				users[account] = struct{}{}
			}
		}
	}

	m.EXPECT().GetUserBySigner(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, signer address.Address) (auth.ListUsersResponse, error) {
		accounts, ok := signers[signer]
		if !ok {
			return nil, errors.New("not exist")
		}
		users := make(auth.ListUsersResponse, 0)
		for account := range accounts {
			users = append(users, &auth.OutputUser{Name: account})
		}
		return users, nil
	}).AnyTimes()

	m.EXPECT().SignerExistInUser(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user string, signer address.Address) (bool, error) {
		accounts, ok := signers[signer]
		if !ok {
			return false, nil
		}
		if _, ok := accounts[user]; ok {
			return true, nil
		}
		return false, nil
	}).AnyTimes()

	m.EXPECT().ListSigners(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, user string) (auth.ListSignerResp, error) {
		addrs := make([]*auth.OutputSigner, 0)
		for signer := range signers {
			if _, ok := signers[signer][user]; ok {
				addrs = append(addrs, &auth.OutputSigner{Signer: signer})
			}
		}
		return addrs, nil
	}).AnyTimes()
}

func (m *AuthClient) UpsertMiner(ctx context.Context, user string, miner string, openMining bool) (bool, error) {
	panic("implement me")
}

func NewMockAuthClient(t *testing.T) *AuthClient {
	ctrl := gomock.NewController(t)
	return &AuthClient{
		MockIAuthClient: mocks.NewMockIAuthClient(ctrl),
	}
}

var _ jwtclient.IAuthClient = (*AuthClient)(nil)
