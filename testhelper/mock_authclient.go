package testhelper

import (
	"errors"
	"sync"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

type AuthClient struct {
	// key: signer address
	signers  map[string]map[string]struct{}
	lkSigner sync.RWMutex
}

func (m *AuthClient) GetUser(req *auth.GetUserRequest) (*auth.OutputUser, error) {
	panic("Don't call me")
}

func (m *AuthClient) GetUserByMiner(req *auth.GetUserByMinerRequest) (*auth.OutputUser, error) {
	panic("Don't call me")
}

func (m *AuthClient) GetUserBySigner(signer string) (auth.ListUsersResponse, error) {
	m.lkSigner.Lock()
	defer m.lkSigner.Unlock()

	accounts, ok := m.signers[signer]
	if !ok {
		return nil, errors.New("not exist")
	}

	users := make(auth.ListUsersResponse, 0)
	for account := range accounts {
		users = append(users, &auth.OutputUser{Name: account})
	}

	return users, nil
}

func (m *AuthClient) RegisterSigners(userName string, signers []string) error {
	panic("Don't call me")
}

func (m *AuthClient) UnregisterSigners(userName string, signers []string) error {
	panic("Don't call me")
}

func (m *AuthClient) VerifyUsers(names []string) error {
	panic("Don't call me")
}

func (m *AuthClient) AddMockUserAndSigner(account string, addrs []address.Address) {
	m.lkSigner.Lock()
	defer m.lkSigner.Unlock()

	for _, signer := range addrs {
		if signer.Protocol() == address.ID {
			signer, _ = ResolveIDAddr(signer)
		}

		users, ok := m.signers[signer.String()]
		if !ok {
			newUsers := make(map[string]struct{})
			newUsers[account] = struct{}{}
			m.signers[signer.String()] = newUsers
		} else {
			if _, ok := users[account]; !ok {
				users[account] = struct{}{}
			}
		}
	}
}

func NewMockAuthClient() *AuthClient {
	return &AuthClient{
		signers: make(map[string]map[string]struct{}),
	}
}

var _ jwtclient.IAuthClient = (*AuthClient)(nil)
