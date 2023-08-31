package priceposter

import (
	"context"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func sendTx(
	ctx context.Context,
	keyBase keyring.Keyring,
	authClient Auth,
	txClient TxService,
	feeder sdk.AccAddress,
	txConfig client.TxConfig,
	ir codectypes.InterfaceRegistry,
	chainID string,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {
	// get key from keybase, can't fail
	keyInfo, err := keyBase.KeyByAddress(feeder)
	if err != nil {
		panic(err)
	}

	// set msgs, can't fail
	txBuilder := txConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		panic(err)
	}

	// get fee, can fail
	fee, gasLimit, err := getFee(msgs)
	if err != nil {
		return nil, err
	}

	txBuilder.SetFeeAmount(fee)
	txBuilder.SetGasLimit(gasLimit)

	// get acc info, can fail
	accNum, sequence, err := getAccount(ctx, authClient, ir, feeder)
	if err != nil {
		return nil, err
	}

	txFactory := tx.Factory{}.
		WithChainID(chainID).
		WithKeybase(keyBase).
		WithTxConfig(txConfig).
		WithAccountNumber(accNum).
		WithSequence(sequence)

	// sign tx, can't fail
	err = tx.Sign(txFactory, keyInfo.Name, txBuilder, true)
	if err != nil {
		panic(err)
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		panic(err)
	}

	resp, err := txClient.BroadcastTx(ctx, &txservice.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		return nil, err
	}
	if resp.TxResponse.Code != abcitypes.CodeTypeOK {
		return resp.TxResponse, fmt.Errorf("tx failed: %s", resp.TxResponse.RawLog)
	}
	return resp.TxResponse, nil
}

func getAccount(ctx context.Context, authClient Auth, ir codectypes.InterfaceRegistry, feeder sdk.AccAddress) (uint64, uint64, error) {
	accRaw, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: feeder.String()})
	if err != nil {
		return 0, 0, err // if account not found it's pointless to continue
	}

	var acc authtypes.AccountI
	err = ir.UnpackAny(accRaw.Account, &acc)
	if err != nil {
		panic(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}

func getFee(_ []sdk.Msg) (sdk.Coins, uint64, error) {
	return sdk.NewCoins(sdk.NewInt64Coin("unibi", 25_000)), 1_000_000, nil
}
