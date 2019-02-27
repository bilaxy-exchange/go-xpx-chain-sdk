// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by a BSD-style
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
package integration

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/proximax-storage/go-xpx-catapult-sdk/sdk"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
	"golang.org/x/net/context"
	"math/big"
	math "math/rand"
	"testing"
	"time"
)

const timeout = 2 * time.Minute
const networkType = sdk.MijinTest
const privateKey = "EE5FABC70136EE4C729F63FE330AA62073D2C6AA65983E6C6170B74A2EC4DD13"

var defaultAccount, _ = sdk.NewAccountFromPrivateKey(privateKey, networkType)

var cfg, _ = sdk.NewConfig(testUrl, networkType)
var ctx = context.Background()

var client = sdk.NewClient(nil, cfg)
var ws, _ = sdk.NewConnectWs(testUrl, 15000000)

type CreateTransaction func() (sdk.Transaction, error)

func sendTransaction(t *testing.T, createTransaction CreateTransaction, account *sdk.Account) {
	//// The confirmedAdded channel notifies when a transaction related to an
	//// address is included in a block. The message contains the transaction.
	chConfirmedAdded, _ := ws.Subscribe.ConfirmedAdded(account.Address)

	tx, err := createTransaction()

	signTx, err := account.Sign(tx)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	_, err = client.Transaction.Announce(ctx, signTx)
	assert.Nil(t, err)

	select {
	case data := <-chConfirmedAdded.Ch:
		fmt.Printf("ConfirmedAdded Tx Content: %v \n", data.GetAbstractTransaction().Hash)
	case <-time.After(timeout):
		t.Error("Timeout request")
	}

	fmt.Println("Successful!")
}

func TestMosaicDefinitionTransaction(t *testing.T) {
	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewMosaicDefinitionTransaction(
			sdk.NewDeadline(time.Hour),
			nonce,
			defaultAccount.PublicAccount.PublicKey,
			sdk.NewMosaicProperties(true, true, true, 4, big.NewInt(1)),
			networkType)
	}, defaultAccount)
}

func TestTransferTransaction(t *testing.T) {
	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewTransferTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.NewAddress("SDUP5PLHDXKBX3UU5Q52LAY4WYEKGEWC6IB3VBFM", networkType),
			[]*sdk.Mosaic{},
			sdk.NewPlainMessage("Test"),
			networkType,
		)

	}, defaultAccount)
}

func TestModifyMultisigTransaction(t *testing.T) {
	acc1, err := sdk.NewAccountFromPublicKey("68b3fbb18729c1fde225c57f8ce080fa828f0067e451a3fd81fa628842b0b763", networkType)
	assert.Nilf(t, err, "NewAccountFromPublicKey returned error: %s", err)
	acc2, err := sdk.NewAccountFromPublicKey("cf893ffcc47c33e7f68ab1db56365c156b0736824a0c1e273f9e00b8df8f01eb", networkType)
	assert.Nilf(t, err, "NewAccountFromPublicKey returned error: %s", err)

	multisigAccount, err := sdk.NewAccount(sdk.MijinTest)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewModifyMultisigAccountTransaction(
			sdk.NewDeadline(time.Hour),
			2,
			1,
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1,
				},
				{
					sdk.Add,
					acc2,
				},
			},
			networkType,
		)
	}, multisigAccount)
}

func TestModifyContracTransaction(t *testing.T) {
	acc1, err := sdk.NewAccountFromPublicKey("68b3fbb18729c1fde225c57f8ce080fa828f0067e451a3fd81fa628842b0b763", networkType)
	assert.Nilf(t, err, "NewAccountFromPublicKey returned error: %s", err)
	acc2, err := sdk.NewAccountFromPublicKey("cf893ffcc47c33e7f68ab1db56365c156b0736824a0c1e273f9e00b8df8f01eb", networkType)
	assert.Nilf(t, err, "NewAccountFromPublicKey returned error: %s", err)

	contractAccount, err := sdk.NewAccount(sdk.MijinTest)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewModifyContractTransaction(
			sdk.NewDeadline(time.Hour),
			2,
			"cf893ffcc47c33e7f68ab1db56365c156b0736824a0c1e273f9e00b8df8f01eb",
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1,
				},
				{
					sdk.Add,
					acc2,
				},
			},
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1,
				},
				{
					sdk.Add,
					acc2,
				},
			},
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1,
				},
				{
					sdk.Add,
					acc2,
				},
			},
			networkType,
		)
	}, contractAccount)
}

func TestRegisterRootNamespaceTransaction(t *testing.T) {
	name := make([]byte, 5)

	_, err := rand.Read(name)
	assert.Nil(t, err)
	nameHex := hex.EncodeToString(name)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewRegisterRootNamespaceTransaction(
			sdk.NewDeadline(time.Hour),
			nameHex,
			big.NewInt(1),
			networkType,
		)
	}, defaultAccount)
}

func TestLockFundsTransactionTransaction(t *testing.T) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	assert.Nil(t, err)
	hash := sdk.Hash(hex.EncodeToString(key))

	stx := &sdk.SignedTransaction{sdk.AggregateBonded, "payload", hash}
	//id, err := sdk.NewMosaicId(big.NewInt(0x20B5A75C59C18264))
	//assert.Nil(t, err)
	//mosaic, err := sdk.NewMosaic(id, big.NewInt(10000000))
	//assert.Nil(t, err)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewLockFundsTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.XemRelative(10),
			big.NewInt(100),
			stx,
			networkType,
		)
	}, defaultAccount)
}

func TestSecretTransactionTransaction(t *testing.T) {
	proof := make([]byte, 8)

	_, err := rand.Read(proof)
	assert.Nil(t, err)

	result := sha3.New256()
	_, err = result.Write(proof)
	assert.Nil(t, err)

	secret := hex.EncodeToString(result.Sum(nil))

	recipient := defaultAccount.PublicAccount.Address

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewSecretLockTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.XemRelative(10),
			big.NewInt(100),
			sdk.SHA3_256,
			secret,
			recipient,
			networkType,
		)
	}, defaultAccount)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewSecretProofTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.SHA3_256,
			secret,
			hex.EncodeToString(proof),
			networkType,
		)
	}, defaultAccount)
}

func TestCompleteAggregateTransactionTransaction(t *testing.T) {
	ttx, err := sdk.NewTransferTransaction(
		sdk.NewDeadline(time.Hour),
		sdk.NewAddress("SBILTA367K2LX2FEXG5TFWAS7GEFYAGY7QLFBYKC", networkType),
		[]*sdk.Mosaic{},
		sdk.NewPlainMessage("test-message"),
		networkType,
	)
	assert.Nil(t, err)
	ttx.ToAggregate(defaultAccount.PublicAccount)

	sendTransaction(t, func() (sdk.Transaction, error) {
		return sdk.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{ttx},
			networkType,
		)
	}, defaultAccount)
}