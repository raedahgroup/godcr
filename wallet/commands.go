package wallet

import (
	"errors"

	"github.com/raedahgroup/dcrlibwallet"
)

var (
	// ErrInvalidArguments is returned when a wallet command is send with invalid arguments.
	ErrInvalidArguments = errors.New("invalid command arguments")

	// ErrNotFound is returned when a wallet command is given that does not exist or is not
	// implemented.
	ErrNotFound = errors.New("command not found or not implemented")

	// ErrNoSuchWallet is returned with the wallet requested by the given id does not exist
	ErrNoSuchWallet = errors.New("no such wallet with id")

	// ErrNoSuchAcct is returned when the given account number cannot be found
	ErrNoSuchAcct = errors.New("no such account")

	// ErrCreateTx is returned when a tx author cannot be created
	ErrCreateTx = errors.New("can not create transaction")
)

// CreateWallet creates a new wallet with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) CreateWallet(passphrase string) {
	go func() {
		var resp Response
		wall, err := wal.multi.CreateNewWallet(passphrase, dcrlibwallet.PassphraseTypePass)
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		resp.Resp = &CreatedSeed{
			Seed: wall.Seed,
		}
		wal.Send <- resp
	}()
}

// RestoreWallet restores a wallet with the given parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) RestoreWallet(seed, passphrase string) {
	go func() {
		var resp Response
		_, err := wal.multi.RestoreWallet(seed, passphrase, dcrlibwallet.PassphraseTypePass)
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		resp.Resp = &Restored{}
		wal.Send <- resp
	}()
}

// CreateTransaction creates a TxAuthor with the given parameters.
// The created TxAuthor will have to have a destination added before broadcasting.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) CreateTransaction(walletID int, accountID int32) {
	go func() {
		var resp Response
		wallets, err := wal.wallets()
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		var wallet *dcrlibwallet.Wallet
		for i := range wallets {
			if wallets[i].ID == walletID {
				wallet = wallets[i]
				break
			}
		}

		if wallet == nil {
			resp.Err = ErrNoSuchWallet
			wal.Send <- resp
			return
		}

		if _, err := wallet.GetAccount(accountID, dcrlibwallet.DefaultRequiredConfirmations); err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		txAuthor := wallet.NewUnsignedTx(accountID, dcrlibwallet.DefaultRequiredConfirmations)
		if txAuthor == nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		resp.Resp = txAuthor
		wal.Send <- resp
	}()
}

// BroadcastTransaction broadcasts the transaction built with txAuthor to the network.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) BroadcastTransaction(txAuthor *dcrlibwallet.TxAuthor, passphrase string) {
	go func() {
		var resp Response

		txHash, err := txAuthor.Broadcast([]byte(passphrase))
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}

		resp.Resp = &TxHash{
			Hash: string(txHash),
		}
		wal.Send <- resp
	}()
}

// GetAllTransactions collects a per-wallet slice of transactions fitting the parameters.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) GetAllTransactions(offset, limit, txfilter int32) {
	go func() {
		var resp Response
		wallets, err := wal.wallets()
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		alltxs := make([][]dcrlibwallet.Transaction, len(wallets))
		for i, wall := range wallets {
			txs, err := wall.GetTransactionsRaw(offset, limit, txfilter, true)
			if err != nil {
				resp.Err = err
				wal.Send <- resp
				return
			}
			alltxs[i] = txs
		}

		resp.Resp = &Transactions{
			Txs: alltxs,
		}
		wal.Send <- resp
	}()
}

// GetMultiWalletInfo gets bulk information about the loaded wallets.
// Information regarding transactions is collected with respect to wal.confirms as the
// number of required confirmations for said transactions.
// It is non-blocking and sends its result or any error to wal.Send.
func (wal *Wallet) GetMultiWalletInfo() {
	go func() {
		var resp Response
		wallets, err := wal.wallets()
		if err != nil {
			resp.Err = err
			wal.Send <- resp
			return
		}
		var completeTotal int64
		infos := make([]InfoShort, len(wallets))
		for i, wall := range wallets {
			iter, err := wall.AccountsIterator(wal.confirms)
			if err != nil {
				resp.Err = err
				wal.Send <- resp
				return
			}

			walletAccounts := []Account{}
			var totalWalletBalance, spendableWalletBalance int64
			for acct := iter.Next(); acct != nil; acct = iter.Next() {
				totalWalletBalance += acct.TotalBalance
				spendableWalletBalance += acct.Balance.Spendable

				account := Account{
					Number:           acct.Number,
					Name:             acct.Name,
					TotalBalance:     acct.Balance.Total,
					SpendableBalance: acct.Balance.Spendable,
				}
				walletAccounts = append(walletAccounts, account)
			}

			completeTotal += totalWalletBalance
			infos[i] = InfoShort{
				ID:               wall.ID,
				Name:             wall.Name,
				TotalBalance:     totalWalletBalance,
				SpendableBalance: spendableWalletBalance,
				Accounts:         walletAccounts,
			}
		}
		best := wal.multi.GetBestBlock()

		if best == nil {
			resp.Err = InternalWalletError{
				Message: "Could not get load best block",
			}
			wal.Send <- resp
			return
		}

		resp.Resp = &MultiWalletInfo{
			LoadedWallets:   len(wallets),
			TotalBalance:    completeTotal,
			BestBlockHeight: best.Height,
			BestBlockTime:   best.Timestamp,
			Wallets:         infos,
			Synced:          wal.multi.IsSynced(),
		}
		wal.Send <- resp
	}()
}

// RenameWallet renames the wallet identified by walletID.
func (wal *Wallet) RenameWallet(walletID int, name string) error {
	return wal.multi.RenameWallet(walletID, name)
}

// CurrentAddress returns the next address for the specified wallet account.
func (wal *Wallet) CurrentAddress(walletID int, accountID int32) (string, error) {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return "", ErrNoSuchWallet
	}
	return wall.CurrentAddress(accountID)
}

// NextAddress returns the next address for the specified wallet account.
func (wal *Wallet) NextAddress(walletID int, accountID int32) (string, error) {
	wall := wal.multi.WalletWithID(walletID)
	if wall == nil {
		return "", ErrNoSuchWallet
	}
	return wall.NextAddress(accountID)
}

// IsAddressValid checks if the given address is valid for the multiwallet network
func (wal *Wallet) IsAddressValid(address string) (bool, error) {
	wall := wal.multi.FirstOrDefaultWallet()
	if wall == nil {
		return false, &InternalWalletError{
			Message: "No wallet loaded",
		}
	}
	return wall.IsAddressValid(address), nil
}

// StartSync starts the multiwallet SPV sync
func (wal *Wallet) StartSync() error {
	return wal.multi.SpvSync()
}

// CancelSync cancels the SPV sync
func (wal *Wallet) CancelSync() {
	go wal.multi.CancelSync()
}
