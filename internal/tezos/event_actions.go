package tezos

import (
	"context"

	"github.com/hyperledger-firefly/tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// EventStreamStart starts an event stream with an initial set of listeners (which might be empty), a channel to deliver events, and a context that will close to stop the stream
func (c *tezosConnector) EventStreamStart(ctx context.Context, req *ffcapi.EventStreamStartRequest) (*ffcapi.EventStreamStartResponse, ffcapi.ErrorReason, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	es := c.eventStreams[*req.ID]
	if es != nil {
		return nil, ffcapi.ErrorReason(""), i18n.NewError(ctx, msgs.MsgStreamAlreadyStarted, req.ID)
	}

	es = &eventStream{
		id:             req.ID,
		c:              c,
		ctx:            req.StreamContext,
		events:         req.EventStream,
		headBlock:      -1,
		listeners:      make(map[fftypes.UUID]*listener),
		streamLoopDone: make(chan struct{}),
	}

	chainHead := c.blockListener.getHighestBlock(ctx)
	for _, lReq := range req.InitialListeners {
		l, err := es.addEventListener(ctx, lReq)
		if err != nil {
			return nil, "", err
		}
		// During initial start we move the "head" block forwards to be the highest of all the initial streams
		if l.hwmBlock > es.headBlock {
			if l.hwmBlock > chainHead {
				es.headBlock = chainHead
			} else {
				es.headBlock = l.hwmBlock
			}
		}
	}

	// From this point we consider ourselves started
	c.eventStreams[*req.ID] = es

	// Start all the listeners
	for _, l := range es.listeners {
		es.startEventListener(l)
	}

	// Start the listener head routine, which reads events for all listeners that are not in catchup mode
	go es.streamLoop()

	// Add the block consumer
	c.blockListener.addConsumer(&blockUpdateConsumer{
		id:      es.id,
		ctx:     req.StreamContext,
		updates: req.BlockListener,
	})

	return &ffcapi.EventStreamStartResponse{}, "", nil
}

// EventStreamStopped informs a connector that an event stream has been requested to stop, and the context has been cancelled. So the state associated with it can be removed (and a future start of the same ID can be performed)
func (c *tezosConnector) EventStreamStopped(ctx context.Context, req *ffcapi.EventStreamStoppedRequest) (*ffcapi.EventStreamStoppedResponse, ffcapi.ErrorReason, error) {
	c.mux.Lock()
	es := c.eventStreams[*req.ID]
	c.mux.Unlock()
	if es != nil {
		select {
		case <-es.ctx.Done():
			// This is good, it is stopped
		default:
			return nil, ffcapi.ErrorReason(""), i18n.NewError(ctx, msgs.MsgStreamNotStopped, req.ID)
		}

		c.mux.Lock()
		delete(c.eventStreams, *req.ID)
		listeners := make([]*listener, 0)
		for _, l := range es.listeners {
			listeners = append(listeners, l)
		}
		c.mux.Unlock()
		// Wait for stream loop to complete
		<-es.streamLoopDone
		// Wait for any listener catchup loops
		for _, l := range listeners {
			if l.catchupLoopDone != nil {
				<-l.catchupLoopDone
			}
		}
	}
	return &ffcapi.EventStreamStoppedResponse{}, "", nil
}

// EventListenerVerifyOptions validates the configuration options for a listener, applying any defaults needed by the connector, and returning the update options for FFTM to persist
func (c *tezosConnector) EventListenerVerifyOptions(_ context.Context, _ *ffcapi.EventListenerVerifyOptionsRequest) (*ffcapi.EventListenerVerifyOptionsResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerAdd begins/resumes listening on set of events that must be consistently ordered. Blockchain specific signatures of the events are included, along with initial conditions (initial block number etc.), and the last stored checkpoint (if any)
func (c *tezosConnector) EventListenerAdd(_ context.Context, _ *ffcapi.EventListenerAddRequest) (*ffcapi.EventListenerAddResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerRemove ends listening on a set of events previous started
func (c *tezosConnector) EventListenerRemove(_ context.Context, _ *ffcapi.EventListenerRemoveRequest) (*ffcapi.EventListenerRemoveResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventListenerHWM queries the current high water mark checkpoint for a listener. Called at regular intervals when there are no events in flight for a listener, to ensure checkpoint are written regularly even when there is no activity
func (c *tezosConnector) EventListenerHWM(_ context.Context, _ *ffcapi.EventListenerHWMRequest) (*ffcapi.EventListenerHWMResponse, ffcapi.ErrorReason, error) {
	return nil, "", nil
}

// EventStreamNewCheckpointStruct used during checkpoint restore, to get the specific into which to restore the JSON bytes
func (c *tezosConnector) EventStreamNewCheckpointStruct() ffcapi.EventListenerCheckpoint {
	return nil
}
