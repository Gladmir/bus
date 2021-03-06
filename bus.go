// Package bus, acts as a messaging "bus" with complete transporting support for protocol buffer objects.
// It handles the framing, reliability of connections, throttling  and  ordered delivery of messages.
//
// Further documentation and samples can be found at https://github.com/gladmir/bus
package bus

import (
	"errors"
	"fmt"
	"sync"
)

type PromiseState string

const (
	SendScheduled  = "SendScheduled"
	Queued = "Queued"
	Cancelled = "Cancelled"
	Sent = "Sent"
	FailedTimeout = "FailedTimeout"
	FailedSerialization = "FailedSerialization"
	FailedTransport = "FailedTransport"

	BusError_InvalidTransport = errors.New("Invalid transport")
	BusError_DestInfoMissing = errors.New("Missing fqdn/ip or port")
	BusError_EndpointAlreadyRegistered = errors.New("Endpoint already registered")
	BusError_MissingPrototypeInstance = errors.New("Prototype instance is missing")
	BusError_MissingMessageHandler = errors.New("Missing message handler implementation")

)

var (
	// contexts map holds the references of 'live' contexts with keys generated via
	// [endpoint's id] + [endpoint's ip] + [endpoint's port] + [endpoints's transport]
	contexts map[string]*socketContext = make(map[string]*socketContext)

	// Similar to contexts (client side contexts), servedContexts holds the references of 'live' contexts currently served.
	servedContexts map[string]*socketContext = make(map[string]*socketContext)

	// key = endpointId, value net.Listener.
	endpointListeners map[string]*listenerShutdown = make(map[string]*listenerShutdown)

	cLock sync.RWMutex
	scLock sync.RWMutex
	elLock sync.RWMutex

)

// Dial, Idiomatic entry point for client side communication to socket endpoints.
func Dial(e *Endpoint) (Context, error) {

	destAddress, cKey, err := evalAddressAndKey(e)

	if err != nil {
		return nil, err
	}

	in, out := createBuckets(e.ThrottlingCriteria)

	ctx := &socketContext{
		e:            e,
		resolvedDest: destAddress,
		ctxId:        cKey,
		ctxQueue:     make(chan *busPromise, e.BufferSize),
		ctxQuit:      make(chan struct{}),
		netQuit:      make(chan struct{}),
		inBucket:in,
		outBucket:out,
		strategy:e.ThrottlingCriteria.Strategy,
	}

	ctx.setState(Opening)

	err = dial(e.Transport, destAddress, ctx)

	if err != nil {
		return nil, err
	}

	cLock.Lock()
	contexts[cKey] = ctx
	cLock.Unlock()

	return ctx, nil
}

// Serve, Idiomatic function for server side communication via Bus
//
// As opposed to Dial, Serve provides the contexts attached to given endpoints within ContextHandler
// 'Opening' callback for the first time, due to the nature of being server side.
//
// Usually, server side coding resides within finite state machines attached to individual messaging contexts which in our case,
// ContextHandler and MessageHandler implementations. Practically, if you really need to implement logic out of these two interfaces,
// just hold on to Context within ContextHandler's 'Opening' callback and use the context wherever you see fit (similar to client side usage).
// Like client side bus usage, ContextHandler implementation is still optional.
//
// TL;DR:
//
// If you need to send a message to connected clients as soon as they arrive, implement the ContextHandler and MessageHandler
//
// If connected clients will start the messaging and depending on the logic you want to provide if you only need to respond to client requests,
// just ignore the ContextHandler implementation, MessageHandler implementation is all you need.
//
// See StopServing and StopServingAll functions for stopping server side endpoints without a context handle.
//
// Serve func never blocks and returns the initial errors, if any, via ResultFunc func callback in a separate goroutine.
// You can pass nil for ResultFunc func parameter if you don't want an error callback.
func Serve(r ResultFunc, e ...*Endpoint) {

	for _, v := range e {

		err := serve(v)

		if err != nil && r != nil {
			go r(v, err)
		}
	}
}

// ResultFunc for initialization and shutdown callback for endpoints on server side bus.
type ResultFunc func(e *Endpoint, err error)

// StopServing, Stops endpoint(s)
// Open contexts will be closed and finally corresponding server side endpoints will be shutdown.
// Only errors which happen during the endpoint shutdowns will be reported via ResultFunc.
//
// As usual, pass nil ResultFunc if you don't want/interested an/in error callback.
func StopServing(r ResultFunc, endpoints ...*Endpoint) {

	for _, e := range endpoints {

		l := endpointListeners[e.Id]

		if l == nil {
			if r != nil {
				r(e, errors.New(fmt.Sprintf("Endpoint already stopped or never served with id: %s", e.Id)))
			}
		}

		// close the contexts first

		for _, c := range servedContexts {
			if c.Endpoint().Id == e.Id {
				c.Close()
			}
		}

		l.q <- struct{}{}
		l.l.Close()
	}
}

// StopServingAll, Stops all contexts and endpoints currently live and running in bus.
func StopServingAll() {

	for _, c := range servedContexts {
		c.Close()
	}

	for _, l := range endpointListeners {
		l.q <- struct{}{}
		l.l.Close()
	}
}

// Stop, Short hand func for;
//  - Close all client and server contexts
//  - Close all served endpoints
func Stop(r ResultFunc) {

	for _, cc := range contexts {
		cc.Close()
	}

	StopServingAll()
}

func createBuckets(t *ThrottlingCriteria) (in Bucket, out Bucket) {

	if nil == t {
		return nil
	}

	in = newLeakyBucket(t.IncomingLimitPerSecond, t.IncomingLimitPerSecond)
	out = newLeakyBucket(t.OutgoingLimitPerSecond, t.OutgoingLimitPerSecond)

	return
}
