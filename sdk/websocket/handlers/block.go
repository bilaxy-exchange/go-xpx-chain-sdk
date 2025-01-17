package handlers

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk"
	"github.com/bilaxy-exchange/go-xpx-chain-sdk/sdk/websocket/subscribers"
)

func NewBlockHandler(messageMapper sdk.BlockMapper, handlers subscribers.Block) *blockHandler {
	return &blockHandler{
		messageMapper: messageMapper,
		handlers:      handlers,
	}
}

type blockHandler struct {
	messageMapper sdk.BlockMapper
	handlers      subscribers.Block
}

func (h *blockHandler) Handle(address *sdk.Address, resp []byte) bool {
	res, err := h.messageMapper.MapBlock(resp)
	if err != nil {
		panic(errors.Wrap(err, "message mapping"))
	}

	handlers := h.handlers.GetHandlers()
	if len(handlers) == 0 {
		return true
	}

	var wg sync.WaitGroup

	for f := range handlers {
		wg.Add(1)
		go func(f *subscribers.BlockHandler) {
			defer wg.Done()

			callFunc := *f

			if rm := callFunc(res); !rm {
				return
			}

			_, err = h.handlers.RemoveHandlers(f)
			if err != nil {
				panic(errors.Wrap(err, "removing handler from storage"))
			}

			return
		}(f)
	}

	wg.Wait()

	return h.handlers.HasHandlers()
}
