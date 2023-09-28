package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	targetspb "github.com/hashicorp/boundary/sdk/pbs/controller/api/resources/targets"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-secure-stdlib/base62"
	"github.com/mr-tron/base58"
	"google.golang.org/protobuf/proto"
)

const sessionCancelTimeout = 10 * time.Second

type ClientProxy struct {
	tofuToken               string
	cachedListenerAddrPort  *atomic.Pointer[netip.AddrPort]
	connectionsLeft         *atomic.Int32
	connsLeftCh             chan int32
	callerConnectionsLeftCh chan int32
	sessionAuthzData        *targetspb.SessionAuthorizationData
	createTime              time.Time
	expiration              time.Time
	ctx                     context.Context
	cancel                  context.CancelFunc
	transport               *http.Transport
	workerAddr              string
	listenAddrPort          netip.AddrPort
	listener                *atomic.Pointer[net.TCPListener]
	listenerCloseOnce       *sync.Once
	clientTlsConf           *tls.Config
	connWg                  *sync.WaitGroup
}

// New creates a new client proxy. The given context should be cancelable; once
// the proxy is started, cancel the context to stop the proxy. The proxy may
// also cancel on its own if the session expires or there are no connections
// left.
func New(ctx context.Context, authzToken string, opt ...Option) (*ClientProxy, error) {
	opts, err := GetOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("could not parse options: %w", err)
	}

	p := &ClientProxy{
		cachedListenerAddrPort:  new(atomic.Pointer[netip.AddrPort]),
		connsLeftCh:             make(chan int32),
		connectionsLeft:         new(atomic.Int32),
		listener:                new(atomic.Pointer[net.TCPListener]),
		listenerCloseOnce:       new(sync.Once),
		connWg:                  new(sync.WaitGroup),
		listenAddrPort:          opts.WithListenAddr,
		callerConnectionsLeftCh: opts.WithConnectionsLeftCh,
	}

	p.tofuToken, err = base62.Random(20)
	if err != nil {
		return nil, fmt.Errorf("could not derive random bytes for tofu token: %w", err)
	}

	marshaled, err := base58.FastBase58Decoding(authzToken)
	if err != nil {
		return nil, fmt.Errorf("unable to base58-decode authorization token: %w", err)
	}
	if len(marshaled) == 0 {
		return nil, errors.New("zero-length authorization information after decoding")
	}
	p.sessionAuthzData = new(targetspb.SessionAuthorizationData)
	if err := proto.Unmarshal(marshaled, p.sessionAuthzData); err != nil {
		return nil, fmt.Errorf("unable to unmarshal authorization data: %w", err)
	}
	if len(p.sessionAuthzData.WorkerInfo) == 0 {
		return nil, errors.New("no workers found in authorization data")
	}

	if p.listenAddrPort.Port() == 0 {
		p.listenAddrPort = netip.AddrPortFrom(p.listenAddrPort.Addr(), uint16(p.sessionAuthzData.DefaultClientPort))
	}
	p.connectionsLeft.Store(p.sessionAuthzData.ConnectionLimit)
	p.workerAddr = p.sessionAuthzData.WorkerInfo[0].Address

	tlsConf, err := p.clientTlsConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating TLS configuration: %w", err)
	}
	p.createTime = p.sessionAuthzData.CreatedTime.AsTime()
	p.expiration = tlsConf.Certificates[0].Leaf.NotAfter

	// We don't _rely_ on client-side timeout verification but this prevents us
	// seeming to be ready for a connection that will immediately fail when we
	// try to actually make it
	p.ctx, p.cancel = context.WithDeadline(ctx, p.expiration)

	transport := cleanhttp.DefaultTransport()
	transport.DisableKeepAlives = false
	// This isn't/shouldn't used anyways really because the connection is
	// hijacked, just setting for completeness
	transport.IdleConnTimeout = 0
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &tls.Dialer{Config: tlsConf}
		return dialer.DialContext(ctx, network, addr)
	}
	p.transport = transport

	return p, nil
}

// Start starts the listener for client proxying. It ends, with any errors, when
// the listener is closed and no connections are left. Cancel the client's proxy
// to force this to happen early. Once call exits, it is not safe to call Start
// again; create a new ClientProxy with New().
func (p *ClientProxy) Start() (retErr error) {
	defer p.cancel()

	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   p.listenAddrPort.Addr().AsSlice(),
		Port: int(p.listenAddrPort.Port()),
	})
	if err != nil {
		return fmt.Errorf("unable to start listening: %w", err)
	}
	p.listener.Store(ln)

	listenerCloseFunc := func() {
		// Forces the for loop to exit instead of spinning on errors
		p.connectionsLeft.Store(0)
		if err := p.listener.Load().Close(); err != nil && err != net.ErrClosed {
			retErr = errors.Join(retErr, fmt.Errorf("error closing proxy listener: %w", err))
		}
	}

	// Ensure closing the listener runs on any other return condition
	defer func() {
		p.listenerCloseOnce.Do(listenerCloseFunc)
	}()

	p.connWg.Add(1)
	go func() {
		defer p.connWg.Done()
		for {
			listeningConn, err := p.listener.Load().AcceptTCP()
			if err != nil {
				select {
				case <-p.ctx.Done():
					return
				default:
					if err == net.ErrClosed {
						// Generally this will be because we canceled the
						// context or ran out of session connections and are
						// winding down. This will never revert, so return.
						return
					}
					// TODO: Log/alert in some way?
					continue
				}
			}
			p.connWg.Add(1)
			go func() {
				defer listeningConn.Close()
				defer p.connWg.Done()
				wsConn, err := p.getWsConn(p.ctx)
				if err != nil {
					// TODO: Log/alert in some way?
					return
				}
				if err := p.runTcpProxyV1(wsConn, listeningConn); err != nil {
					// TODO: Log/alert in some way?
				}
			}()
		}
	}()

	p.connWg.Add(1)
	go func() {
		defer func() {
			// Run a function (last, after connwg is done) just to ensure that
			// we drain from this in case any connections starting as this
			// number changes are trying to send the information down
			for {
				select {
				case <-p.connsLeftCh:
				default:
					return
				}
			}
		}()
		defer p.connWg.Done()
		defer p.listenerCloseOnce.Do(listenerCloseFunc)

		for {
			select {
			case <-p.ctx.Done():
				return
			case connsLeft := <-p.connsLeftCh:
				p.connectionsLeft.Store(connsLeft)
				if p.callerConnectionsLeftCh != nil {
					p.callerConnectionsLeftCh <- connsLeft
				}
				// TODO: Surface this to caller
				if connsLeft == 0 {
					// Close the listener as we can't authorize any more
					// connections
					return
				}
			}
		}
	}()

	p.connWg.Wait()

	// Teardown. Only do it if we haven't expired or reached connection limit
	// since we don't need to clean up in that case.
	var sendSessionCancel bool
	select {
	case <-p.ctx.Done():
		sendSessionCancel = true
	default:
		// If we're not after expiration, ensure there is a bit of buffer in
		// case clocks are not quite the same between worker and this machine
		if time.Now().Before(p.expiration.Add(-5 * time.Minute)) {
			sendSessionCancel = true
		}
	}

	if !sendSessionCancel {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sessionCancelTimeout)
	defer cancel()
	if err := p.sendSessionTeardown(ctx); err != nil {
		return fmt.Errorf("error sending session teardown request to worker: %w", err)
	}

	return nil
}

// ListenerAddr returns the address of the client proxy listener. Because the
// listener is started with Start(), this could be called before listening
// occurs. To avoid returning until we have a valid value, pass a context;
// canceling the context, or passing a nil context when the listener has not yet
// been started, will cause the function to return an empty AddrPort. Otherwise
// the function will return when the address is available. In either case, test
// the result with IsValid.
//
// Warning: a non-cancelable context will cause this call to block forever until
// the listener's address can be determined.
func (p *ClientProxy) ListenerAddr(ctx context.Context) netip.AddrPort {
	switch {
	case p.cachedListenerAddrPort.Load() != nil:
		return *p.cachedListenerAddrPort.Load()
	case p.listener.Load() != nil:
		addrPort := p.listener.Load().Addr().(*net.TCPAddr).AddrPort()
		p.cachedListenerAddrPort.Store(&addrPort)
		return addrPort
	case ctx == nil:
		return netip.AddrPort{}
	}
	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return netip.AddrPort{}
		case <-timer.C:
			if p.listener.Load() != nil {
				timer.Stop()
				addrPort := p.listener.Load().Addr().(*net.TCPAddr).AddrPort()
				p.cachedListenerAddrPort.Store(&addrPort)
				return addrPort
			}
			timer.Reset(10 * time.Millisecond)
		}
	}
}

// SessionCreation returns the creation time of the session
func (p *ClientProxy) SessionCreation() time.Time {
	return p.createTime
}

// SessionExpiration returns the expiration time of the session
func (p *ClientProxy) SessionExpiration() time.Time {
	return p.expiration
}

// ConnectionsLeft returns the number of connections left in the session, or -1
// if unlimited
func (p *ClientProxy) ConnectionsLeft() int32 {
	return p.connectionsLeft.Load()
}
