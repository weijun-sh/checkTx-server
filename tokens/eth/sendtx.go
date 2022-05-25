package eth

import (
	"errors"

	"github.com/weijun-sh/rsyslog/log"
	"github.com/weijun-sh/rsyslog/params"
	"github.com/weijun-sh/rsyslog/types"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	tx, ok := signedTx.(*types.Transaction)
	if !ok {
		log.Printf("signed tx is %+v", signedTx)
		return "", errors.New("wrong signed transaction type")
	}
	txHash, err = b.SendSignedTransaction(tx)
	if err != nil {
		log.Info("SendTransaction failed", "hash", txHash, "err", err)
	} else {
		log.Info("SendTransaction success", "hash", txHash)
		if !params.IsParallelSwapEnabled() {
			sender, errt := types.Sender(b.Signer, tx)
			if errt != nil {
				log.Error("SendTransaction get sender failed", "tx", txHash, "err", errt)
				return txHash, errt
			}
			b.SetNonce(sender.LowerHex(), tx.Nonce()+1)
		}
	}
	if params.IsDebugMode() {
		log.Infof("SendTransaction rawtx is %v", tx.RawStr())
	}
	return txHash, err
}
