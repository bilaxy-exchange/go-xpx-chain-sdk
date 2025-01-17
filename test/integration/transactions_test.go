// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by a BSD-style
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
package integration

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	math "math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk"
)

const defaultDurationNamespaceAndMosaic = 1000000

var listening = false

type CreateTransaction func() (sdk.Transaction, error)

type Result struct {
	sdk.Transaction
	error
}

func initListeners(t *testing.T, account *sdk.Account, hash *sdk.Hash) <-chan Result {
	if !listening {
		// Starting listening messages from websocket
		go wsc.Listen()
		listening = true
	}

	out := make(chan Result)

	// Register handlers functions for needed topics
	if err := wsc.AddConfirmedAddedHandlers(account.Address, func(transaction sdk.Transaction) bool {
		if !hash.Equal(transaction.GetAbstractTransaction().TransactionHash) {
			return false
		}
		fmt.Printf("ConfirmedAdded Tx Content: %v \n", transaction)
		fmt.Println("Successful!")
		out <- Result{transaction, nil}
		return true
	}); err != nil {
		panic(err)
	}

	if err := wsc.AddStatusHandlers(account.Address, func(info *sdk.StatusInfo) bool {
		if !hash.Equal(info.Hash) {
			return false
		}
		fmt.Printf("Got error: %v \n", info)
		t.Error()
		out <- Result{nil, errors.New(info.Status)}
		return true
	}); err != nil {
		panic(err)
	}

	return out
}

func waitTimeout(t *testing.T, wg <-chan Result, timeout time.Duration) Result {
	select {
	case result := <-wg:
		return result
	case <-time.After(timeout):
		t.Error("Timeout request")
		return Result{nil, errors.New("Timeout error")}
	}
}

func sendTransaction(t *testing.T, createTransaction CreateTransaction, account *sdk.Account) Result {
	tx, err := createTransaction()
	assert.Nil(t, err)

	signTx, err := account.Sign(tx)
	assert.Nil(t, err)

	assert.Nil(t, err)
	wg := initListeners(t, account, signTx.Hash)
	_, err = client.Transaction.Announce(ctx, signTx)
	assert.Nil(t, err)

	return waitTimeout(t, wg, timeout)
}

func sendAggregateTransaction(t *testing.T, createTransaction func() (*sdk.AggregateTransaction, error), account *sdk.Account, cosignatories ...*sdk.Account) Result {
	tx, err := createTransaction()
	assert.Nil(t, err)

	signTx, err := account.SignWithCosignatures(tx, cosignatories)
	assert.Nil(t, err)

	stx := &sdk.SignedTransaction{sdk.AggregateBonded, "payload", signTx.Hash}

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewLockFundsTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.XpxRelative(10),
			sdk.Duration(100),
			stx,
		)
	}, account)

	if result.error != nil {
		return result
	}

	time.Sleep(2 * time.Second)

	wg := initListeners(t, account, signTx.Hash)
	_, err = client.Transaction.AnnounceAggregateBonded(ctx, signTx)
	assert.Nil(t, err)

	return waitTimeout(t, wg, timeout)
}

func TestAccountLinkTransaction(t *testing.T) {
	rootAccount, err := client.NewAccount()
	assert.Nil(t, err)
	fmt.Println(rootAccount)
	childAccount, err := client.NewAccount()
	assert.Nil(t, err)
	fmt.Println(childAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewAccountLinkTransaction(
			sdk.NewDeadline(time.Hour),
			childAccount.PublicAccount,
			sdk.AccountLink,
		)
	}, rootAccount)
	assert.Nil(t, result.error)
}

func TestNetworkConfigTransaction(t *testing.T) {
	config, err := client.Network.GetNetworkConfig(ctx)
	assert.Nil(t, err)

	prevValue := config.NetworkConfig.Sections["plugin:catapult.plugins.upgrade"].Fields["minUpgradePeriod"].Value
	config.NetworkConfig.Sections["plugin:catapult.plugins.upgrade"].Fields["minUpgradePeriod"].Value = "1"

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewNetworkConfigTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.Duration(1),
			config.NetworkConfig,
			config.SupportedEntityVersions)
	}, nemesisAccount)
	assert.Nil(t, result.error)

	time.Sleep(time.Minute)

	config.NetworkConfig.Sections["plugin:catapult.plugins.upgrade"].Fields["minUpgradePeriod"].Value = prevValue
	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewNetworkConfigTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.Duration(5),
			config.NetworkConfig,
			config.SupportedEntityVersions)
	}, nemesisAccount)
	assert.Nil(t, result.error)
}

// This test will break blockchain, so only for local testing
func TestBlockchainUpgradeTransaction(t *testing.T) {
	network, err := client.Network.GetNetworkVersion(ctx)
	assert.Nil(t, err)
	version := network.BlockChainVersion + 1

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewBlockchainUpgradeTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.Duration(361),
			version)
	}, nemesisAccount)
	assert.Nil(t, result.error)
}

func TestMosaicDefinitionTransaction(t *testing.T) {
	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewMosaicDefinitionTransaction(
			sdk.NewDeadline(time.Hour),
			nonce,
			defaultAccount.PublicAccount.PublicKey,
			sdk.NewMosaicProperties(true, true, 4, sdk.Duration(defaultDurationNamespaceAndMosaic)),
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestMosaicDefinitionTransaction_ZeroDuration(t *testing.T) {
	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewMosaicDefinitionTransaction(
			sdk.NewDeadline(time.Hour),
			nonce,
			defaultAccount.PublicAccount.PublicKey,
			sdk.NewMosaicProperties(true, true, 4, sdk.Duration(0)),
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestTransferTransaction(t *testing.T) {
	recipientAccount, err := client.NewAccount()
	assert.Nil(t, err)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewTransferTransaction(
			sdk.NewDeadline(time.Hour),
			recipientAccount.Address,
			[]*sdk.Mosaic{},
			sdk.NewPlainMessage("Test"),
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestTransferTransaction_SecureMessage(t *testing.T) {
	const message = "I love you forever"
	recipientAccount, err := client.NewAccount()
	assert.Nil(t, err)

	secureMessage, err := sdk.NewSecureMessageFromPlaintText(message, defaultAccount.PrivateKey, recipientAccount.KeyPair.PublicKey)
	assert.Nil(t, err)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewTransferTransaction(
			sdk.NewDeadline(time.Hour),
			recipientAccount.PublicAccount.Address,
			[]*sdk.Mosaic{},
			secureMessage,
		)
	}, defaultAccount)
	assert.Nil(t, result.error)

	transfer := result.Transaction.(*sdk.TransferTransaction)
	plainMessage, err := recipientAccount.DecryptMessage(
		transfer.Message.(*sdk.SecureMessage),
		defaultAccount.PublicAccount,
	)

	assert.Equal(t, message, plainMessage.Message())
}

func TestModifyMultisigTransaction(t *testing.T) {
	acc1, err := client.NewAccount()
	assert.Nil(t, err)
	acc2, err := client.NewAccount()
	assert.Nil(t, err)

	multisigAccount, err := client.NewAccount()
	assert.Nil(t, err)
	fmt.Println(multisigAccount)

	multTxs, err := client.NewModifyMultisigAccountTransaction(
		sdk.NewDeadline(time.Hour),
		2,
		1,
		[]*sdk.MultisigCosignatoryModification{
			{
				sdk.Add,
				acc1.PublicAccount,
			},
			{
				sdk.Add,
				acc2.PublicAccount,
			},
		},
	)
	assert.Nil(t, err)
	multTxs.ToAggregate(multisigAccount.PublicAccount)

	fackeTxs, err := client.NewTransferTransaction(
		sdk.NewDeadline(time.Hour),
		multisigAccount.PublicAccount.Address,
		[]*sdk.Mosaic{},
		sdk.NewPlainMessage("I wan't to create multisig"),
	)
	assert.Nil(t, err)
	fackeTxs.ToAggregate(defaultAccount.PublicAccount)

	result := sendAggregateTransaction(t, func() (*sdk.AggregateTransaction, error) {
		return client.NewBondedAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{multTxs, fackeTxs},
		)
	}, defaultAccount, multisigAccount, acc1, acc2)
	assert.Nil(t, result.error)
}

func TestModifyContracTransaction(t *testing.T) {
	acc1, err := client.NewAccount()
	assert.Nil(t, err)
	acc2, err := client.NewAccount()
	assert.Nil(t, err)
	hash, err := sdk.StringToHash("cf893ffcc47c33e7f68ab1db56365c156b0736824a0c1e273f9e00b8df8f01eb")
	assert.Nil(t, err)

	contractAccount, err := client.NewAccount()
	fmt.Println(contractAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewModifyContractTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.Duration(2),
			hash,
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1.PublicAccount,
				},
				{
					sdk.Add,
					acc2.PublicAccount,
				},
			},
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1.PublicAccount,
				},
				{
					sdk.Add,
					acc2.PublicAccount,
				},
			},
			[]*sdk.MultisigCosignatoryModification{
				{
					sdk.Add,
					acc1.PublicAccount,
				},
				{
					sdk.Add,
					acc2.PublicAccount,
				},
			},
		)
	}, contractAccount)
	assert.Nil(t, result.error)
}

func TestRegisterRootNamespaceTransaction(t *testing.T) {
	name := make([]byte, 5)

	_, err := rand.Read(name)
	assert.Nil(t, err)
	nameHex := hex.EncodeToString(name)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewRegisterRootNamespaceTransaction(
			sdk.NewDeadline(time.Hour),
			nameHex,
			sdk.Duration(defaultDurationNamespaceAndMosaic),
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestLockFundsTransactionTransaction(t *testing.T) {
	hash := &sdk.Hash{}

	_, err := rand.Read(hash[:])
	assert.Nil(t, err)

	stx := &sdk.SignedTransaction{sdk.AggregateBonded, "payload", hash}

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewLockFundsTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.XpxRelative(10),
			sdk.Duration(100),
			stx,
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestSecretTransaction(t *testing.T) {
	for _, hashType := range []sdk.HashType{sdk.SHA_256, sdk.KECCAK_256, sdk.SHA3_256, sdk.HASH_160} {
		proofB := make([]byte, 8)
		_, err := rand.Read(proofB)
		assert.Nil(t, err)

		proof := sdk.NewProofFromBytes(proofB)
		secret, err := proof.Secret(hashType)
		assert.Nil(t, err)
		recipient := defaultAccount.PublicAccount.Address

		result := sendTransaction(t, func() (sdk.Transaction, error) {
			return client.NewSecretLockTransaction(
				sdk.NewDeadline(time.Hour),
				sdk.XpxRelative(10),
				sdk.Duration(100),
				secret,
				recipient,
			)
		}, defaultAccount)
		assert.Nil(t, result.error)

		result = sendTransaction(t, func() (sdk.Transaction, error) {
			return client.NewSecretProofTransaction(
				sdk.NewDeadline(time.Hour),
				hashType,
				proof,
				recipient,
			)
		}, defaultAccount)
		assert.Nil(t, result.error)
	}
}

func TestCompleteAggregateTransaction(t *testing.T) {
	acc, err := client.NewAccount()
	assert.Nil(t, err)

	ttx, err := client.NewTransferTransaction(
		sdk.NewDeadline(time.Hour),
		acc.Address,
		[]*sdk.Mosaic{},
		sdk.NewPlainMessage("test-message"),
	)
	assert.Nil(t, err)
	ttx.ToAggregate(defaultAccount.PublicAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{ttx},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestAggregateBoundedTransaction(t *testing.T) {
	receiverAccount, err := client.NewAccount()

	ttx1, err := client.NewTransferTransaction(
		sdk.NewDeadline(time.Hour),
		receiverAccount.Address,
		[]*sdk.Mosaic{},
		sdk.NewPlainMessage("test-message"),
	)
	assert.Nil(t, err)
	ttx1.ToAggregate(defaultAccount.PublicAccount)

	ttx2, err := client.NewTransferTransaction(
		sdk.NewDeadline(time.Hour),
		defaultAccount.Address,
		[]*sdk.Mosaic{},
		sdk.NewPlainMessage("test-message"),
	)
	assert.Nil(t, err)
	ttx2.ToAggregate(receiverAccount.PublicAccount)

	result := sendAggregateTransaction(t, func() (*sdk.AggregateTransaction, error) {
		return client.NewBondedAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{ttx1, ttx2},
		)
	}, defaultAccount, receiverAccount)
	assert.Nil(t, result.error)
}

func TestAddressAliasTransaction(t *testing.T) {
	name := make([]byte, 5)

	_, err := rand.Read(name)
	assert.Nil(t, err)
	nameHex := hex.EncodeToString(name)

	nsId, err := sdk.NewNamespaceIdFromName(nameHex)
	assert.Nil(t, err)

	registerTx, err := client.NewRegisterRootNamespaceTransaction(
		sdk.NewDeadline(time.Hour),
		nameHex,
		sdk.Duration(defaultDurationNamespaceAndMosaic),
	)
	assert.Nil(t, err)
	registerTx.ToAggregate(defaultAccount.PublicAccount)

	aliasTx, err := client.NewAddressAliasTransaction(
		sdk.NewDeadline(time.Hour),
		defaultAccount.PublicAccount.Address,
		nsId,
		sdk.AliasLink,
	)
	assert.Nil(t, err)
	aliasTx.ToAggregate(defaultAccount.PublicAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{registerTx, aliasTx},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)

	senderAccount, err := client.NewAccount()
	assert.Nil(t, err)

	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewTransferTransactionWithNamespace(
			sdk.NewDeadline(time.Hour),
			nsId,
			[]*sdk.Mosaic{},
			sdk.NewPlainMessage("Test"),
		)
	}, senderAccount)
	assert.Nil(t, result.error)
}

func TestMosaicAliasTransaction(t *testing.T) {
	name := make([]byte, 5)

	_, err := rand.Read(name)
	assert.Nil(t, err)
	nameHex := hex.EncodeToString(name)

	nsId, err := sdk.NewNamespaceIdFromName(nameHex)
	assert.Nil(t, err)

	registerTx, err := client.NewRegisterRootNamespaceTransaction(
		sdk.NewDeadline(time.Hour),
		nameHex,
		sdk.Duration(defaultDurationNamespaceAndMosaic),
	)
	assert.Nil(t, err)
	registerTx.ToAggregate(defaultAccount.PublicAccount)

	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	mosaicId, err := sdk.NewMosaicIdFromNonceAndOwner(nonce, defaultAccount.PublicAccount.PublicKey)
	assert.Nil(t, err)
	mosaicDefinitionTx, err := client.NewMosaicDefinitionTransaction(
		sdk.NewDeadline(time.Hour),
		nonce,
		defaultAccount.PublicAccount.PublicKey,
		sdk.NewMosaicProperties(true, true, 4, sdk.Duration(defaultDurationNamespaceAndMosaic)),
	)
	assert.Nil(t, err)
	mosaicDefinitionTx.ToAggregate(defaultAccount.PublicAccount)

	aliasTx, err := client.NewMosaicAliasTransaction(
		sdk.NewDeadline(time.Hour),
		mosaicId,
		nsId,
		sdk.AliasLink,
	)
	assert.Nil(t, err)
	aliasTx.ToAggregate(defaultAccount.PublicAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{registerTx, mosaicDefinitionTx, aliasTx},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestModifyAddressMetadataTransaction(t *testing.T) {
	fmt.Println(defaultAccount.PublicAccount.Address)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewModifyMetadataAddressTransaction(
			sdk.NewDeadline(time.Hour),
			defaultAccount.PublicAccount.Address,
			[]*sdk.MetadataModification{
				{
					sdk.AddMetadata,
					"jora229",
					"I Love you!",
				},
			})
	}, defaultAccount)
	assert.Nil(t, result.error)

	time.Sleep(2 * time.Second)

	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewModifyMetadataAddressTransaction(
			sdk.NewDeadline(time.Hour),
			defaultAccount.PublicAccount.Address,
			[]*sdk.MetadataModification{
				{
					sdk.RemoveMetadata,
					"jora229",
					"",
				},
			})
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestModifyMosaicMetadataTransaction(t *testing.T) {
	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	mosaicDefinitionTx, err := client.NewMosaicDefinitionTransaction(
		sdk.NewDeadline(time.Hour),
		nonce,
		defaultAccount.PublicAccount.PublicKey,
		sdk.NewMosaicProperties(true, true, 4, sdk.Duration(defaultDurationNamespaceAndMosaic)),
	)
	assert.Nil(t, err)
	mosaicDefinitionTx.ToAggregate(defaultAccount.PublicAccount)

	mosaicId, err := sdk.NewMosaicIdFromNonceAndOwner(nonce, defaultAccount.PublicAccount.PublicKey)
	assert.Nil(t, err)

	fmt.Println(mosaicId.String())

	metadataTx, err := client.NewModifyMetadataMosaicTransaction(
		sdk.NewDeadline(time.Hour),
		mosaicId,
		[]*sdk.MetadataModification{
			{
				sdk.AddMetadata,
				"hello",
				"hell",
			},
		},
	)
	assert.Nil(t, err)
	metadataTx.ToAggregate(defaultAccount.PublicAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{mosaicDefinitionTx, metadataTx},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)

	time.Sleep(2 * time.Second)

	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewModifyMetadataMosaicTransaction(
			sdk.NewDeadline(time.Hour),
			mosaicId,
			[]*sdk.MetadataModification{
				{
					sdk.RemoveMetadata,
					"hello",
					"",
				},
			},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestModifyNamespaceMetadataTransaction(t *testing.T) {
	name := make([]byte, 5)

	_, err := rand.Read(name)
	assert.Nil(t, err)
	nameHex := hex.EncodeToString(name)

	namespaceId, err := sdk.NewNamespaceIdFromName(nameHex)
	assert.Nil(t, err)
	fmt.Println(namespaceId)

	registrNamespaceTx, err := client.NewRegisterRootNamespaceTransaction(
		sdk.NewDeadline(time.Hour),
		nameHex,
		sdk.Duration(defaultDurationNamespaceAndMosaic),
	)
	assert.Nil(t, err)
	registrNamespaceTx.ToAggregate(defaultAccount.PublicAccount)

	modifyMetadataTx, err := client.NewModifyMetadataNamespaceTransaction(
		sdk.NewDeadline(time.Hour),
		namespaceId,
		[]*sdk.MetadataModification{
			{
				sdk.AddMetadata,
				"hello",
				"world",
			},
		},
	)
	assert.Nil(t, err)
	modifyMetadataTx.ToAggregate(defaultAccount.PublicAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewCompleteAggregateTransaction(
			sdk.NewDeadline(time.Hour),
			[]sdk.Transaction{registrNamespaceTx, modifyMetadataTx},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)

	time.Sleep(2 * time.Second)

	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewModifyMetadataNamespaceTransaction(
			sdk.NewDeadline(time.Hour),
			namespaceId,
			[]*sdk.MetadataModification{
				{
					sdk.RemoveMetadata,
					"hello",
					"",
				},
			},
		)
	}, defaultAccount)
	assert.Nil(t, result.error)
}

func TestAccountPropertiesAddressTransaction(t *testing.T) {
	blockAccount, err := client.NewAccount()
	assert.Nil(t, err)
	testAccount, err := client.NewAccount()
	assert.Nil(t, err)

	fmt.Println(blockAccount, testAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewAccountPropertiesAddressTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.BlockAddress,
			[]*sdk.AccountPropertiesAddressModification{
				{
					sdk.AddProperty,
					blockAccount.Address,
				},
			},
		)
	}, testAccount)
	assert.Nil(t, result.error)
}

func TestAccountPropertiesMosaicTransaction(t *testing.T) {
	r := math.New(math.NewSource(time.Now().UTC().UnixNano()))
	nonce := r.Uint32()

	mosaicId, err := sdk.NewMosaicIdFromNonceAndOwner(nonce, defaultAccount.PublicAccount.PublicKey)
	assert.Nil(t, err)

	fmt.Println(mosaicId.String())

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewMosaicDefinitionTransaction(
			sdk.NewDeadline(time.Hour),
			nonce,
			defaultAccount.PublicAccount.PublicKey,
			sdk.NewMosaicProperties(true, true, 4, sdk.Duration(defaultDurationNamespaceAndMosaic)),
		)
	}, defaultAccount)
	assert.Nil(t, result.error)

	testAccount, err := client.NewAccount()
	assert.Nil(t, err)

	fmt.Println(testAccount)

	result = sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewAccountPropertiesMosaicTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.BlockMosaic,
			[]*sdk.AccountPropertiesMosaicModification{
				{
					sdk.AddProperty,
					mosaicId,
				},
			},
		)
	}, testAccount)
	assert.Nil(t, result.error)
}

func TestAccountPropertiesEntityTypeTransaction(t *testing.T) {
	testAccount, err := client.NewAccount()
	assert.Nil(t, err)

	fmt.Println(testAccount)

	result := sendTransaction(t, func() (sdk.Transaction, error) {
		return client.NewAccountPropertiesEntityTypeTransaction(
			sdk.NewDeadline(time.Hour),
			sdk.BlockTransaction,
			[]*sdk.AccountPropertiesEntityTypeModification{
				{
					sdk.AddProperty,
					sdk.LinkAccount,
				},
			},
		)
	}, testAccount)
	assert.Nil(t, result.error)
}
