// Code generated by mockery v1.0.0. DO NOT EDIT.

// Copyright 2019 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package mocks

import mock "github.com/stretchr/testify/mock"

// Router is an autogenerated mock type for the Router type
type Router struct {
	mock.Mock
}

// RouteMessage provides a mock function with given fields: _a0
func (_m *Router) RouteMessage(_a0 []byte) {
	_m.Called(_a0)
}

// SetUid provides a mock function with given fields: _a0
func (_m *Router) SetUid(_a0 string) {
	_m.Called(_a0)
}
