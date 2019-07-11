// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by a BSD-style
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.
package integration

import (
	"context"
	"github.com/proximax-storage/go-xpx-catapult-sdk/sdk"
	"github.com/proximax-storage/go-xpx-catapult-sdk/sdk/websocket"
	"github.com/proximax-storage/go-xpx-utils/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	cfg, err := sdk.NewConfigFromRemote([]string{testUrl})
	if err != nil {
		panic(err)
	}

	ctx = context.Background()
	client = sdk.NewClient(nil, cfg)

	wsc, err = websocket.NewClient(ctx, cfg)
	if err != nil {
		panic(err)
	}

	defaultAccount, err = client.NewAccountFromPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
}

func TestAddressService_GetAccountNames(t *testing.T) {

	networkType := sdk.MijinTest

	addresses := []*sdk.Address{
		{
			networkType,
			"SCWXLOABHP4FT2LWTT3Z6GDCHLLMUIKKFRBE2O3S",
		},
		{
			networkType,
			"SBKDKHFIRM72EAVDT6TI426CKUCP5DQIJV73XB5X",
		},
	}

	names, err := client.Account.GetAccountNames(
		ctx,
		addresses...)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(names), len(addresses))

	for i, accNames := range names {
		tests.ValidateStringers(t, addresses[i], accNames.Address)
	}
}
