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

func Test_confirmedAddedHandler_Handle(t *testing.T) {
	type fields struct {
		messageMapper sdk.ConfirmedAddedMapper
		handlers      subscribers.ConfirmedAdded
	}
	type args struct {
		address *sdk.Address
		resp    []byte
	}

	address := new(sdk.Address)

	obj := new(sdk.TransferTransaction)
	messageMapperMock := new(mappers.ConfirmedAddedMapper)
	messageMapperMock.On("MapConfirmedAdded", mock.Anything).Return(obj, nil)

	handlerFunc1 := func(tx sdk.Transaction) bool {
		return false
	}

	handlerFunc2 := func(tx sdk.Transaction) bool {
		return true
	}

	blockHandler1 := subscribers.ConfirmedAddedHandler(handlerFunc1)
	blockHandler2 := subscribers.ConfirmedAddedHandler(handlerFunc2)

	blockHandlers := map[*subscribers.ConfirmedAddedHandler]struct{}{
		&blockHandler1: {},
		&blockHandler2: {},
	}

	blockHandlersMock := new(mocksSubscribers.ConfirmedAdded)
	blockHandlersMock.On("GetHandlers", mock.Anything).Return(nil).Once().
		On("GetHandlers", mock.Anything).Return(blockHandlers).
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
				handlers:      blockHandlersMock,
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
				handlers:      blockHandlersMock,
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
			h := &confirmedAddedHandler{
				messageMapper: tt.fields.messageMapper,
				handlers:      tt.fields.handlers,
			}
			got := h.Handle(tt.args.address, tt.args.resp)
			assert.Equal(t, got, tt.want)
		})
	}
}
