// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sdk

import (
	"fmt"

	"github.com/proximax-storage/go-xpx-utils/str"
)

type MosaicId struct {
	baseInt64
}

// returns MosaicId for passed mosaic identifier
func NewMosaicId(id uint64) (*MosaicId, error) {
	if hasBits(id, NamespaceBit) {
		return nil, ErrWrongBitMosaicId
	}
	return newMosaicIdPanic(id), nil
}

func newMosaicIdPanic(id uint64) *MosaicId {
	mosaicId := MosaicId{baseInt64(id)}
	return &mosaicId
}

func (m *MosaicId) Type() AssetIdType {
	return MosaicAssetIdType
}

func (m *MosaicId) Id() uint64 {
	return uint64(m.baseInt64)
}

func (m *MosaicId) String() string {
	return m.toHexString()
}

func (m *MosaicId) toHexString() string {
	return uint64ToHex(m.Id())
}

func (m *MosaicId) Equals(id AssetId) bool {
	return m.Id() == id.Id()
}

// returns MosaicId for passed nonce and public key of mosaic owner
func NewMosaicIdFromNonceAndOwner(nonce uint32, ownerPublicKey string) (*MosaicId, error) {
	if len(ownerPublicKey) != 64 {
		return nil, ErrInvalidOwnerPublicKey
	}

	return generateMosaicId(nonce, ownerPublicKey)
}

type Mosaic struct {
	AssetId AssetId
	Amount  Amount
}

// returns a Mosaic for passed AssetId and amount
func NewMosaic(assetId AssetId, amount Amount) (*Mosaic, error) {
	if assetId == nil {
		return nil, ErrNilAssetId
	}

	return newMosaicPanic(assetId, amount), nil
}

// returns a Mosaic for passed AssetId and amount without validation of parameters
func newMosaicPanic(assetId AssetId, amount Amount) *Mosaic {
	return &Mosaic{
		AssetId: assetId,
		Amount:  amount,
	}
}

func (m *Mosaic) String() string {
	return str.StructToString(
		"MosaicId",
		str.NewField("AssetId", str.StringPattern, m.AssetId),
		str.NewField("Amount", str.StringPattern, m.Amount),
	)
}

type MosaicInfo struct {
	MosaicId   *MosaicId
	Supply     Amount
	Height     Height
	Owner      *PublicAccount
	Revision   uint32
	Properties *MosaicProperties
}

func (m *MosaicInfo) String() string {
	return str.StructToString(
		"MosaicInfo",
		str.NewField("MosaicId", str.StringPattern, m.MosaicId),
		str.NewField("Supply", str.StringPattern, m.Supply),
		str.NewField("Height", str.StringPattern, m.Height),
		str.NewField("Owner", str.StringPattern, m.Owner),
		str.NewField("Revision", str.IntPattern, m.Revision),
		str.NewField("Properties", str.StringPattern, m.Properties),
	)
}

const Supply_Mutable = 0x01
const Transferable = 0x02

// structure which includes several properties for defining mosaic
// `SupplyMutable` - is supply of defined mosaic can be changed in future
// `Transferable` - if this property is set to "false", only transfer transactions having the creator as sender or as recipient can transfer mosaics of that type. If set to "true" the mosaics can be transferred to and from arbitrary accounts
// `Divisibility` - divisibility determines up to what decimal place the mosaic can be divided into
// `Duration` - duration in blocks mosaic will be available. After the renew mosaic is inactive and can be renewed
type MosaicPropertiesHeader struct {
	SupplyMutable bool
	Transferable  bool
	Divisibility  uint8
}

type MosaicProperties struct {
	MosaicPropertiesHeader
	OptionalProperties []MosaicProperty
}

type MosaicProperty struct {
	Id    MosaicPropertyId
	Value baseInt64
}

func (mp *MosaicProperty) String() string {
	return str.StructToString(
		"MosaicProperty",
		str.NewField("Id", str.IntPattern, mp.Id),
		str.NewField("Value", str.IntPattern, mp.Value),
	)
}

// returns MosaicProperties from actual values
func NewMosaicProperties(supplyMutable bool, transferable bool, divisibility uint8, duration Duration) *MosaicProperties {
	properties := make([]MosaicProperty, 0)

	if duration != 0 {
		properties = append(properties, MosaicProperty{MosaicPropertyDurationId, duration})
	}

	ref := &MosaicProperties{
		MosaicPropertiesHeader{
			supplyMutable,
			transferable,
			divisibility,
		},
		properties,
	}

	return ref
}

func (mp *MosaicProperties) String() string {
	return str.StructToString(
		"MosaicProperties",
		str.NewField("SupplyMutable", str.BooleanPattern, mp.SupplyMutable),
		str.NewField("Transferable", str.BooleanPattern, mp.Transferable),
		str.NewField("Divisibility", str.IntPattern, mp.Divisibility),
		str.NewField("OptionalProperties", str.ValuePattern, mp.OptionalProperties),
	)
}

func (mp *MosaicProperties) Duration() Duration {
	for _, property := range mp.OptionalProperties {
		if property.Id == MosaicPropertyDurationId {
			return Duration(property.Value)
		}
	}

	return 0
}

type MosaicName struct {
	MosaicId *MosaicId
	Names    []string
}

func (m *MosaicName) String() string {
	return str.StructToString(
		"MosaicName",
		str.NewField("MosaicId", str.StringPattern, m.MosaicId),
		str.NewField("Names", str.StringPattern, m.Names),
	)
}

type MosaicSupplyType uint8

const (
	Decrease MosaicSupplyType = iota
	Increase
)

func (tx MosaicSupplyType) String() string {
	return fmt.Sprintf("%d", tx)
}

// returns XEM mosaic with passed amount
func Xem(amount uint64) *Mosaic {
	return newMosaicPanic(XemNamespaceId, Amount(amount))
}

// returns XPX mosaic with passed amount
func Xpx(amount uint64) *Mosaic {
	return newMosaicPanic(XpxNamespaceId, Amount(amount))
}

// returns XEM with actual passed amount
func XemRelative(amount uint64) *Mosaic {
	return Xem(1000000 * amount)
}

// returns XPX with actual passed amount
func XpxRelative(amount uint64) *Mosaic {
	return Xpx(1000000 * amount)
}
