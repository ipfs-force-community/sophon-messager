package service

import (
	"context"
	"time"

	"github.com/filecoin-project/go-jsonrpc"

	"gorm.io/gorm"

	"github.com/filecoin-project/go-address"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/types"
)

var ErrRecordExist = xerrors.Errorf("record exist")

type WalletService struct {
	repo           repo.Repo
	log            *logrus.Logger
	cfg            *config.WalletConfig
	sps            *SharedParamsService
	nodeClient     *NodeClient
	addressService *AddressService

	pendingAddrChan   chan pendingAddr
	pendingWalletChan chan pendingWallet
}

type pendingWallet struct {
	walletName string
	walletID   types.UUID
}

type pendingAddr struct {
	walletName string
	addr       address.Address

	walletID, addrID types.UUID
}

func NewWalletService(repo repo.Repo,
	logger *logrus.Logger,
	nodeClient *NodeClient,
	addressService *AddressService,
	cfg *config.WalletConfig,
	sps *SharedParamsService) *WalletService {
	ws := &WalletService{
		repo:           repo,
		log:            logger,
		nodeClient:     nodeClient,
		addressService: addressService,
		cfg:            cfg,
		sps:            sps,

		pendingWalletChan: make(chan pendingWallet, 10),
		pendingAddrChan:   make(chan pendingAddr, 10),
	}
	go ws.listenWalletChange(context.TODO())
	go ws.checkWalletState()
	go ws.checkAddressState()

	return ws
}

func (walletService *WalletService) SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error) {
	// try connect wallet
	_, close, err := NewWalletClient(ctx, wallet.Url, wallet.Token)
	if err != nil {
		return types.UUID{}, err
	}
	close()
	if err := walletService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.WalletRepo().HasWallet(wallet.Name)
		if err != nil {
			return err
		}
		if has {
			w, err := txRepo.WalletRepo().GetOneRecord(wallet.Name)
			if err != nil {
				return err
			}
			if w.IsDeleted == repo.NotDeleted && w.State == types.Alive {
				return ErrRecordExist
			}
			wallet.ID = w.ID
		}
		return txRepo.WalletRepo().SaveWallet(wallet)
	}); err != nil {
		return types.UUID{}, err
	}
	walletService.log.Infof("save wallet %v", wallet)

	return wallet.ID, nil
}

func (walletService *WalletService) GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByID(uuid)
}

func (walletService *WalletService) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return walletService.repo.WalletRepo().GetWalletByName(name)
}

func (walletService *WalletService) HasWallet(ctx context.Context, name string) (bool, error) {
	return walletService.repo.WalletRepo().HasWallet(name)
}

func (walletService *WalletService) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return walletService.repo.WalletRepo().ListWallet()
}

func (walletService *WalletService) GetWalletClient(ctx context.Context, walletName string) (WalletClient, jsonrpc.ClientCloser, error) {
	wallet, err := walletService.GetWalletByName(ctx, walletName)
	if err != nil {
		return WalletClient{}, nil, err
	}
	return NewWalletClient(ctx, wallet.Url, wallet.Token)
}

func (walletService *WalletService) ListRemoteWalletAddress(ctx context.Context, walletName string) ([]address.Address, error) {
	cli, close, err := walletService.GetWalletClient(ctx, walletName)
	if err != nil {
		return nil, err
	}
	defer close()

	return cli.WalletList(ctx)
}

func (walletService *WalletService) DeleteWallet(ctx context.Context, walletName string) (string, error) {
	w, err := walletService.GetWalletByName(ctx, walletName)
	if err != nil {
		return "", err
	}

	if err := walletService.repo.WalletRepo().UpdateState(walletName, types.Removing); err != nil {
		return "", err
	}

	was, err := walletService.repo.WalletAddressRepo().GetWalletAddressByWalletID(w.ID)
	if err != nil {
		return "", err
	}

	for _, wa := range was {
		addr, err := walletService.repo.AddressRepo().GetAddressByID(ctx, wa.AddrID)
		if err != nil {
			walletService.log.Infof("found address(%s) %v", wa.AddrID, err)
			continue
		}
		walletService.delAddress(pendingAddr{walletName: walletName, addr: addr.Addr, walletID: wa.WalletID, addrID: wa.AddrID})
	}
	walletService.pendingWalletChan <- pendingWallet{walletName: walletName, walletID: w.ID}
	walletService.log.Infof("delete wallet %s", walletName)

	return walletName, nil
}

//// wallet address ////

func (walletService *WalletService) HasWalletAddress(ctx context.Context, walletName string, addr address.Address) (bool, error) {
	walletID, addID, err := walletService.getWalletIDAndAddrID(ctx, walletName, addr)
	if err != nil {
		return false, err
	}
	return walletService.repo.WalletAddressRepo().HasWalletAddress(walletID, addID)
}

func (walletService *WalletService) getWalletIDAndAddrID(ctx context.Context, walletName string, addr address.Address) (types.UUID, types.UUID, error) {
	wallet, err := walletService.repo.WalletRepo().GetWalletByName(walletName)
	if err != nil {
		return types.UUID{}, types.UUID{}, xerrors.Errorf("got wallet %v", err)
	}
	addrInfo, err := walletService.repo.AddressRepo().GetAddress(ctx, addr)
	if err != nil {
		return types.UUID{}, types.UUID{}, err
	}

	return wallet.ID, addrInfo.ID, nil
}

func (walletService *WalletService) SetSelectMsgNum(ctx context.Context, walletName string, addr address.Address, num uint64) (address.Address, error) {
	walletID, addID, err := walletService.getWalletIDAndAddrID(ctx, walletName, addr)
	if err != nil {
		return addr, err
	}
	if err := walletService.repo.WalletAddressRepo().UpdateSelectMsgNum(walletID, addID, num); err != nil {
		return addr, err
	}

	return addr, nil
}

func (walletService *WalletService) ListWalletAddress(ctx context.Context) ([]*types.WalletAddress, error) {
	return walletService.repo.WalletAddressRepo().ListWalletAddress()
}

func (walletService *WalletService) GetWalletAddress(ctx context.Context, walletName string, addr address.Address) (*types.WalletAddress, error) {
	walletID, addID, err := walletService.getWalletIDAndAddrID(ctx, walletName, addr)
	if err != nil {
		return nil, err
	}
	return walletService.repo.WalletAddressRepo().GetWalletAddress(walletID, addID)
}

func (walletService *WalletService) processWallet() error {
	walletList, err := walletService.ListWallet(context.TODO())
	if err != nil {
		return err
	}
	for _, w := range walletList {
		cli, close, err := NewWalletClient(context.Background(), w.Url, w.Token)
		if err != nil {
			walletService.log.Errorf("connect wallet(%s) %v", w.Name, err)
			continue
		}
		if err := walletService.syncWalletAddress(context.TODO(), w.Name, w.ID, &cli); err != nil {
			walletService.log.Errorf("process wallet failed %v %v", w.Name, err)
		}
		close()
	}

	return nil
}

func (walletService *WalletService) listenWalletChange(ctx context.Context) {
	interval := time.Duration(walletService.cfg.ScanInterval) * time.Second
	params := walletService.sps.GetParams()
	if params.SharedParams != nil && params.ScanInterval != 0 {
		interval = time.Duration(params.ScanInterval) * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := walletService.processWallet(); err != nil {
				walletService.log.Errorf("process wallet %v", err)
			}
		case i := <-walletService.sps.GetParams().ScanIntervalChan:
			ticker.Reset(i)
		case <-ctx.Done():
			walletService.log.Warnf("context error: %v", ctx.Err())
			return
		}
	}

}

func (walletService *WalletService) syncWalletAddress(ctx context.Context, walletName string, walletID types.UUID, cli IWalletClient) error {
	addrList, err := cli.WalletList(ctx)
	if err != nil {
		return err
	}

	was, err := walletService.repo.WalletAddressRepo().GetWalletAddressByWalletID(walletID)
	if err != nil {
		return xerrors.Errorf("got wallet address by wallet id(%s) %v", walletID, err)
	}
	addrMap := make(map[address.Address]types.UUID, len(was))
	for _, wa := range was {
		addr, err := walletService.repo.AddressRepo().GetAddressByID(context.TODO(), wa.AddrID)
		if err != nil {
			walletService.log.Errorf("got address %v", err)
			continue
		}
		addrMap[addr.Addr] = wa.AddrID
	}
	for _, addr := range addrList {
		if addrID, ok := addrMap[addr]; ok {
			delete(addrMap, addr)
			if wa, err := walletService.repo.WalletAddressRepo().GetWalletAddress(walletID, addrID); err == nil &&
				(wa.AddressState == types.Alive || wa.AddressState == types.Forbiden) {
				continue
			}
		}
		// store address
		addrID, err := walletService.saveAddress(ctx, addr)
		if err != nil {
			walletService.log.Errorf("save address %v", err)
			continue
		}

		if err := walletService.updateWalletAddress(ctx, cli, walletID, addrID); err != nil {
			walletService.log.Errorf("save wallet address %v", err)
			continue
		}
		walletService.log.Infof("wallet %s add address %s", walletName, addr.String())
	}

	// address to handle remote wallet deletion
	for addr, addrID := range addrMap {
		wa, err := walletService.repo.WalletAddressRepo().GetWalletAddress(walletID, addrID)
		if err == nil && wa.AddressState == types.Removing {
			continue
		}
		walletService.delAddress(pendingAddr{walletName: walletName, addr: addr, walletID: walletID, addrID: addrID})
	}

	return nil
}

// update address table
func (walletService *WalletService) saveAddress(ctx context.Context, addr address.Address) (types.UUID, error) {
	var nonce uint64
	actor, err := walletService.nodeClient.StateGetActor(context.Background(), addr, venustypes.EmptyTSK)
	if err != nil {
		walletService.log.Warnf("get actor failed, addr: %s, err: %v", addr, err)
	} else {
		nonce = actor.Nonce //current nonce should big than nonce on chain
	}

	addrTmp := &types.Address{
		ID:        types.NewUUID(),
		Addr:      addr,
		Nonce:     nonce,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		IsDeleted: -1,
	}
	addrID, err := walletService.addressService.SaveAddress(ctx, addrTmp)
	if err != nil && !xerrors.Is(err, ErrRecordExist) {
		return addrID, err
	}

	return addrID, nil
}

// update wallet address table
func (walletService *WalletService) updateWalletAddress(ctx context.Context, cli IWalletClient, walletID, addrID types.UUID) error {
	wa := &types.WalletAddress{
		ID:           types.NewUUID(),
		WalletID:     walletID,
		AddrID:       addrID,
		AddressState: types.Alive,
		IsDeleted:    -1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := walletService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.WalletAddressRepo().HasWalletAddress(walletID, addrID)
		if err != nil {
			return err
		}
		if has {
			walletAddress, err := txRepo.WalletAddressRepo().GetOneRecord(walletID, addrID)
			if err != nil && xerrors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if walletAddress.SelMsgNum != 0 { // inherit SelMsgNum
				wa.SelMsgNum = walletAddress.SelMsgNum
			}
			wa.CreatedAt = walletAddress.CreatedAt
			wa.ID = walletAddress.ID
			if walletAddress.AddressState == types.Forbiden { // keep Forbiden status
				wa.AddressState = walletAddress.AddressState
			}
		}
		return txRepo.WalletAddressRepo().SaveWalletAddress(wa)
	}); err != nil {
		return err
	}

	return nil
}

func (walletService *WalletService) checkWalletState() {
	walletList, err := walletService.repo.WalletRepo().ListWallet()
	if err != nil {
		walletService.log.Errorf("got wallet %v", err)
	}
	for _, w := range walletList {
		if w.State == types.Removing {
			walletService.pendingWalletChan <- pendingWallet{walletName: w.Name, walletID: w.ID}
		}
	}
	for pw := range walletService.pendingWalletChan {
		checkAgain := true
		if wallet, err := walletService.repo.WalletRepo().GetWalletByName(pw.walletName); err == nil && wallet.State == types.Alive {
			checkAgain = false
		}

		was, err := walletService.repo.WalletAddressRepo().GetWalletAddressByWalletID(pw.walletID)
		if err != nil {
			walletService.log.Errorf("got wallet address by wallet id %v", err)
		} else if len(was) == 0 { // All addresses of a wallet have been deleted, then delete wallet
			if err := walletService.repo.WalletRepo().DelWallet(pw.walletName); err != nil {
				walletService.log.Errorf("deleted wallet %v", err)
			} else {
				walletService.log.Infof("deleted wallet %s", pw.walletName)
				checkAgain = false
			}
		}
		if checkAgain {
			go func() {
				time.Sleep(time.Second * 30)
				walletService.pendingWalletChan <- pw
			}()
		}
	}
}

func (walletService *WalletService) checkAddressState() {
	walletAddrList, err := walletService.repo.WalletAddressRepo().ListWalletAddress()
	if err != nil {
		walletService.log.Errorf("get wallet address %v", err)
	}

	for _, wa := range walletAddrList {
		if wa.AddressState == types.Removing {
			addrInfo, err := walletService.repo.AddressRepo().GetAddressByID(context.TODO(), wa.AddrID)
			if err != nil {
				walletService.log.Errorf("found address(%s) %v", wa.AddrID, err)
				continue
			}
			walletInfo, err := walletService.repo.WalletRepo().GetWalletByID(wa.WalletID)
			if err != nil {
				walletService.log.Errorf("found wallet(%s) %v", wa.WalletID, err)
				continue
			}
			walletService.pendingAddrChan <- pendingAddr{walletName: walletInfo.Name, addr: addrInfo.Addr, walletID: wa.WalletID, addrID: wa.AddrID}
		}
	}

	for pa := range walletService.pendingAddrChan {
		var isDeleted bool
		msgs, err := walletService.repo.MessageRepo().ListFilledMessageByWallet(pa.walletName, pa.addr)
		if err != nil {
			walletService.log.Errorf("got filled message %v", err)
		} else if len(msgs) == 0 {
			// add address again
			if wa, err := walletService.repo.WalletAddressRepo().GetWalletAddress(pa.walletID, pa.addrID); err == nil && wa.AddressState == types.Alive {
				isDeleted = true
			} else {
				if err := walletService.repo.WalletAddressRepo().DelWalletAddress(pa.walletID, pa.addrID); err != nil && xerrors.Is(err, gorm.ErrRecordNotFound) {
					walletService.log.Errorf("update address state %v", err)
					continue
				}
				walletService.log.Infof("deleted address %v", pa.addr.String())

				// not using address, delete it
				if has, err := walletService.repo.WalletAddressRepo().HasAddress(pa.addrID); err == nil && !has {
					if _, err = walletService.addressService.DeleteAddress(context.TODO(), pa.addr); err != nil {
						walletService.log.Errorf("delete address(%s) %v", pa.addr.String(), err)
					}
				}
				isDeleted = true
			}
		}
		if !isDeleted {
			go func() {
				time.Sleep(time.Second * 30)
				walletService.pendingAddrChan <- pa
			}()
		}
	}
}

func (walletService *WalletService) ForbiddenAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	walletID, addID, err := walletService.getWalletIDAndAddrID(ctx, walletName, addr)
	if err != nil {
		return address.Undef, err
	}
	if err := walletService.repo.WalletAddressRepo().UpdateAddressState(walletID, addID, types.Forbiden); err != nil {
		return address.Undef, err
	}

	walletService.log.Infof("forbidden address %v", addr.String())

	return addr, nil
}

func (walletService *WalletService) ActiveAddress(ctx context.Context, walletName string, addr address.Address) (address.Address, error) {
	walletID, addID, err := walletService.getWalletIDAndAddrID(ctx, walletName, addr)
	if err != nil {
		return address.Undef, err
	}
	if err := walletService.repo.WalletAddressRepo().UpdateAddressState(walletID, addID, types.Alive); err != nil {
		return address.Undef, err
	}

	walletService.log.Infof("active address %v", addr.String())

	return addr, nil
}

func (walletService *WalletService) AllAddresses() map[address.Address]struct{} {
	addrs := make(map[address.Address]struct{})
	addrList, err := walletService.repo.AddressRepo().ListAddress(context.TODO())
	if err != nil {
		return addrs
	}

	for _, addr := range addrList {
		addrs[addr.Addr] = struct{}{}
	}

	return addrs
}

func (walletService *WalletService) delAddress(pa pendingAddr) {
	if err := walletService.repo.WalletAddressRepo().UpdateAddressState(pa.walletID, pa.addrID, types.Removing); err != nil {
		walletService.log.Errorf("update wallet address state %v", err)
	}
	if err := walletService.repo.MessageRepo().UpdateUnFilledMessageState(pa.walletName, pa.addr, types.NoWalletMsg); err != nil {
		walletService.log.Errorf("update unfilled message state %v", err)
	}
	go func() {
		walletService.pendingAddrChan <- pa
	}()
	walletService.log.Infof("wallet delete address %s", pa.addr.String())
}
