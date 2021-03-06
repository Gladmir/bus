package bus

import (
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
	"time"
)

// Endpoint interface defines the contract for various endpoints for different socket transports.
type Endpoint struct {
	// Id
	//
	// Client side;
	//
	// Optional
	//
	// Id is present for correlation between endpoints and contexts.
	// More practical usage of different endpointIds is when you need to connect to the same endpoint with same ip/port/transport multiple times.
	//
	// Bus differentiates the endpoints by generating keys with;
	//
	//	[Id]-[ip:port]-[transport] for tcp|udp|ws and [Id]-[ip:port]-[resourceUrl] for http|https endpoints
	//
	// Server side;
	//
	// Mandatory
	//
	// Usually frameworks/libraries guard themselves by encapsulating their internal logic by not exposing states which represents
	// uniqueness, Bus is not one of them. If you don't provide an Id or Id() returns nil or returns different values for each call,
	// your application will surely malfunction.
	//
	// So assign an Id and always return the same value for individual endpoint implementations of yours.
	//
	// Package level Stop... functions all depend on EndpointId parameter in order to stop serving endpoints.
	Id                string

	// Address information, ipv4|ipv6
	Address           string

	// Port information
	Port              int

	// FQDN, Fully qualified domain name.
	// Implementors may choose to provide Hostname (FQDN) instead of Address, bus will try to resolve the FQDN if provided.
	FQDN              string

	// Transport may be one of "tcp|udp|ws"
	Transport         string

	// BufferSize, if provided other than zero, defines the message queue size of the endpoint.
	// Bus would still accept messages if Endpoint is not reachable and/or in reconnecting state until endpoint's
	// buffer is full.
	BufferSize        int

	// Used for websockets
	Origin            string

	// The url of the http(s) and websocket endpoints
	ResourceUrl       string

	// reconnect true|false, max attempt count between disconnects, delay between attempts.
	// If you provide zero or negative max attempt count, Bus will try reconnecting forever
	// This method is used only for client side.
	ShouldReconnect   bool
	MaxAttemptCount   int
	DelayDuration     time.Duration

	// PrototypeInstance should return a zero value of User's Protocol Buffer object.
	Prototype         proto.Message

	// Optional, Check documentation of ThrottlingCriteria struct.
	ThrottlingCriteria ThrottlingCriteria

	// Mandatory, Check documentation of MessageHandler interface.
	MessageHandler    MessageHandler

	// Optional, Check documentation of ContextHandler interface.
	ContextHandler    ContextHandler
}

type ThrottlingStrategy string

const (
	BusTs_MPS ThrottlingStrategy = "Message-Per-Second"
	BusTs_BPS ThrottlingStrategy = "Bytes-Per-Second"
)

// Criteria for throttling, optional for all endpoints
// While, MPS strategy should be deterministic, BPS is generally best effort.
//
// Once an endpoint is passed to Bus functions, during live period of the endpoint,
// throttling values are final regardless of different values you return from your implementation
type ThrottlingCriteria struct {
	// MPS or BPS
	Strategy               ThrottlingStrategy

	// BusTS_MPS: Number of incoming messages to be processed per second.
	// BusTS_BPS: Number of incoming bytes to be read per second.
	IncomingLimitPerSecond int

	// BusTS_MPS: Number of outgoing messages to be sent per second.
	// BusTS_BPS: Number of outgoing bytes to be written.
	OutgoingLimitPerSecond int
}

type listenerShutdown struct {
	l net.Listener
	q chan <- struct{}
}

func resolveAddress(e *Endpoint) (string, error) {

	if (e.FQDN == "" && e.Address == "") ||
	e.Port == 0 || e.Transport == "" {

		return "", BusError_DestInfoMissing
	}

	port := strconv.Itoa(e.Port)

	var address string

	// FQDN has a priority over IP
	if e.FQDN != "" {

		// resolve the fqdn
		addrs, err := net.LookupHost(e.FQDN)

		if err != nil && e.Address == "" {
			return "", err
		}

		if len(addrs) == 0 && e.Address == "" {
			return "", BusError_DestInfoMissing
		}

		address = addrs[0]

	} else {
		address = e.Address
	}

	return address + ":" + port, nil

}

func evalAddressAndKey(e *Endpoint) (string, string, error) {

	address, err := resolveAddress(e)

	if err != nil {
		return "", "", err
	}

	cKey := e.Id + "-" + address + "-" + e.Transport

	cLock.RLock()
	defer cLock.RUnlock()
	if contexts[cKey] != nil {
		return "", "", BusError_EndpointAlreadyRegistered
	}

	if e.Prototype == nil {
		return "", "", BusError_MissingPrototypeInstance
	}

	if e.MessageHandler == nil {
		return "", "", BusError_MissingMessageHandler
	}

	return address, cKey, nil

}
