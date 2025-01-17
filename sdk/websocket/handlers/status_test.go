package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bilaxy-exchange/go-xpx-chain-sdk/mocks/mappers"

	mocksSubscribers "github.com/bilaxy-exchange/go-xpx-chain-sdk/mocks/subscribers"
	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk"
	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk/websocket/subscribers"
)

func Test_statusHandler_Handle(t *testing.T) {
	type fields struct {
		messageMapper sdk.StatusMapper
		handlers      subscribers.Status
	}
	type args struct {
		address *sdk.Address
		resp    []byte
	}

	address := new(sdk.Address)

	obj := new(sdk.StatusInfo)
	messageMapperMock := new(mappers.StatusMapper)
	messageMapperMock.On("MapStatus", mock.Anything).Return(obj, nil)

	handlerFunc1 := func(info *sdk.StatusInfo) bool {
		return false
	}

	handlerFunc2 := func(info *sdk.StatusInfo) bool {
		return true
	}

	handler1 := subscribers.StatusHandler(handlerFunc1)
	handler2 := subscribers.StatusHandler(handlerFunc2)

	handlers := map[*subscribers.StatusHandler]struct{}{
		&handler1: {},
		&handler2: {},
	}

	HandlersMock := new(mocksSubscribers.Status)
	HandlersMock.On("GetHandlers", mock.Anything).Return(nil).Once().
		On("GetHandlers", mock.Anything).Return(handlers).
		On("RemoveHandlers", mock.Anything, mock.Anything).Return(true, nil).
		On("HasHandlers", mock.Anything).Return(true, nil)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "empty handlers",
			fields: fields{
				handlers:      HandlersMock,
				messageMapper: messageMapperMock,
			},
			args: args{
				address: address,
			},
			want: true,
		},
		{
			name: "remove handlers without error",
			fields: fields{
				handlers:      HandlersMock,
				messageMapper: messageMapperMock,
			},
			args: args{
				address: address,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &statusHandler{
				messageMapper: tt.fields.messageMapper,
				handlers:      tt.fields.handlers,
			}

			got := h.Handle(tt.args.address, tt.args.resp)
			assert.Equal(t, got, tt.want)
		})
	}
}
