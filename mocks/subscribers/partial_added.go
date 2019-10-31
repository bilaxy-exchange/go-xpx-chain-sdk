// Code generated by mockery v1.0.0. DO NOT EDIT.

// Copyright 2019 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package subscribers

import mock "github.com/stretchr/testify/mock"
import sdk "github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk"
import subscribers "github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk/websocket/subscribers"

// PartialAdded is an autogenerated mock type for the PartialAdded type
type PartialAdded struct {
	mock.Mock
}

// AddHandlers provides a mock function with given fields: address, handlers
func (_m *PartialAdded) AddHandlers(address *sdk.Address, handlers ...subscribers.PartialAddedHandler) error {
	_va := make([]interface{}, len(handlers))
	for _i := range handlers {
		_va[_i] = handlers[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, address)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(*sdk.Address, ...subscribers.PartialAddedHandler) error); ok {
		r0 = rf(address, handlers...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAddresses provides a mock function with given fields:
func (_m *PartialAdded) GetAddresses() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// GetHandlers provides a mock function with given fields: address
func (_m *PartialAdded) GetHandlers(address *sdk.Address) map[*subscribers.PartialAddedHandler]struct{} {
	ret := _m.Called(address)

	var r0 map[*subscribers.PartialAddedHandler]struct{}
	if rf, ok := ret.Get(0).(func(*sdk.Address) map[*subscribers.PartialAddedHandler]struct{}); ok {
		r0 = rf(address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[*subscribers.PartialAddedHandler]struct{})
		}
	}

	return r0
}

// HasHandlers provides a mock function with given fields: address
func (_m *PartialAdded) HasHandlers(address *sdk.Address) bool {
	ret := _m.Called(address)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*sdk.Address) bool); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// RemoveHandlers provides a mock function with given fields: address, handlers
func (_m *PartialAdded) RemoveHandlers(address *sdk.Address, handlers ...*subscribers.PartialAddedHandler) (bool, error) {
	_va := make([]interface{}, len(handlers))
	for _i := range handlers {
		_va[_i] = handlers[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, address)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*sdk.Address, ...*subscribers.PartialAddedHandler) bool); ok {
		r0 = rf(address, handlers...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*sdk.Address, ...*subscribers.PartialAddedHandler) error); ok {
		r1 = rf(address, handlers...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
