package handlers

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk"
	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk/websocket/subscribers"
)

func NewUnconfirmedRemovedHandler(messageMapper sdk.UnconfirmedRemovedMapper, handlers subscribers.UnconfirmedRemoved) *unconfirmedRemovedHandler {
	return &unconfirmedRemovedHandler{
		messageMapper: messageMapper,
		handlers:      handlers,
	}
}

type unconfirmedRemovedHandler struct {
	messageMapper sdk.UnconfirmedRemovedMapper
	handlers      subscribers.UnconfirmedRemoved
}

func (h *unconfirmedRemovedHandler) Handle(address *sdk.Address, resp []byte) bool {
	res, err := h.messageMapper.MapUnconfirmedRemoved(resp)
	if err != nil {
		panic(errors.Wrap(err, "message mapper error"))
	}

	handlers := h.handlers.GetHandlers(address)
	if len(handlers) == 0 {
		return true
	}

	var wg sync.WaitGroup

	for f := range handlers {
		wg.Add(1)
		go func(f *subscribers.UnconfirmedRemovedHandler) {
			defer wg.Done()

			callFunc := *f

			if rm := callFunc(res); !rm {
				return
			}

			_, err = h.handlers.RemoveHandlers(address, f)
			if err != nil {
				panic(errors.Wrap(err, "removing handler from storage"))
			}

			return
		}(f)
	}

	wg.Wait()

	return h.handlers.HasHandlers(address)
}
