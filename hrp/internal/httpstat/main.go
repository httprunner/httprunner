// Package httpstat traces HTTP latency infomation (DNSLookup, TCP Connection and so on) on any golang HTTP request.
// It uses `httptrace` package.
// Forked from https://github.com/tcnksm/go-httpstat
package httpstat

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// Stat stores httpstat info.
type Stat struct {
	// The following are duration for each phase
	// DNSLookup => TCPConnection => TLSHandshake => ServerProcessing => ContentTransfer
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandshake     time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration // from the first response byte to tansfer done.

	// The followings are timeline of request
	NameLookup    time.Duration // = DNSLookup
	Connect       time.Duration // = DNSLookup + TCPConnection
	Pretransfer   time.Duration // = DNSLookup + TCPConnection + TLSHandshake
	StartTransfer time.Duration // = DNSLookup + TCPConnection + TLSHandshake + ServerProcessing
	Total         time.Duration // = DNSLookup + TCPConnection + TLSHandshake + ServerProcessing + ContentTransfer

	// internal timelines, including start and finish timestamps of each phase
	dnsStart      time.Time
	dnsDone       time.Time
	tcpStart      time.Time
	tcpDone       time.Time
	tlsStart      time.Time
	tlsDone       time.Time
	serverStart   time.Time
	serverDone    time.Time
	transferStart time.Time
	transferDone  time.Time // need to be provided from outside

	// isTLS is true when connection seems to use TLS
	isTLS bool

	// isReused is true when connection is reused (keep-alive)
	isReused bool
}

// Finish sets the time when reading response is done.
// This must be called after reading response body.
func (s *Stat) Finish() {
	s.transferDone = time.Now()

	// This means result is empty (it does nothing).
	// Skip setting value (contentTransfer and total will be zero).
	if s.dnsStart.IsZero() {
		return
	}

	s.ContentTransfer = s.transferDone.Sub(s.transferStart)
	s.Total = s.transferDone.Sub(s.dnsStart)
}

// Durations returns all durations and timelines of request latencies
func (s *Stat) Durations() map[string]time.Duration {
	return map[string]time.Duration{
		"DNSLookup":        s.DNSLookup / time.Millisecond,
		"TCPConnection":    s.TCPConnection / time.Millisecond,
		"TLSHandshake":     s.TLSHandshake / time.Millisecond,
		"ServerProcessing": s.ServerProcessing / time.Millisecond,
		"ContentTransfer":  s.ContentTransfer / time.Millisecond,
		"NameLookup":       s.NameLookup / time.Millisecond,
		"Connect":          s.Connect / time.Millisecond,
		"Pretransfer":      s.Connect / time.Millisecond,
		"StartTransfer":    s.StartTransfer / time.Millisecond,
		"Total":            s.Total / time.Millisecond,
	}
}

// WithHTTPStat is a wrapper of httptrace.WithClientTrace.
// It records the time of each httptrace hooks.
func WithHTTPStat(ctx context.Context, s *Stat) context.Context {
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(i httptrace.DNSStartInfo) {
			s.dnsStart = time.Now()
		},

		DNSDone: func(i httptrace.DNSDoneInfo) {
			s.dnsDone = time.Now()

			s.DNSLookup = s.dnsDone.Sub(s.dnsStart)
			s.NameLookup = s.dnsDone.Sub(s.dnsStart)
		},

		ConnectStart: func(_, _ string) {
			s.tcpStart = time.Now()

			// When connecting to IP (When no DNS lookup)
			if s.dnsStart.IsZero() {
				s.dnsStart = s.tcpStart
				s.dnsDone = s.tcpStart
			}
		},

		ConnectDone: func(network, addr string, err error) {
			s.tcpDone = time.Now()
			s.TCPConnection = s.tcpDone.Sub(s.tcpStart)
			s.Connect = s.tcpDone.Sub(s.dnsStart)
		},

		TLSHandshakeStart: func() {
			s.isTLS = true
			s.tlsStart = time.Now()
		},

		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			s.tlsDone = time.Now()
			s.TLSHandshake = s.tlsDone.Sub(s.tlsStart)
			s.Pretransfer = s.tlsDone.Sub(s.dnsStart)
		},

		GotConn: func(i httptrace.GotConnInfo) {
			// Handle when keep alive is used and connection is reused.
			// DNSStart(Done) and ConnectStart(Done) is skipped
			if i.Reused {
				s.isReused = true
			}
		},

		WroteRequest: func(info httptrace.WroteRequestInfo) {
			s.serverStart = time.Now()

			// When client doesn't use DialContext or using old (before go1.7) `net`
			// package, DNS/TCP/TLS hook is not called.
			if s.dnsStart.IsZero() && s.tcpStart.IsZero() {
				now := s.serverStart
				s.dnsStart = now
				s.dnsDone = now
				s.tcpStart = now
				s.tcpDone = now
			}

			// When connection is re-used, DNS/TCP/TLS hook is not called.
			if s.isReused {
				now := s.serverStart
				s.dnsStart = now
				s.dnsDone = now
				s.tcpStart = now
				s.tcpDone = now
				s.tlsStart = now
				s.tlsDone = now
			}

			if s.isTLS {
				return
			}

			s.TLSHandshake = s.tcpDone.Sub(s.tcpDone)
			s.Pretransfer = s.Connect
		},

		GotFirstResponseByte: func() {
			s.serverDone = time.Now()
			s.ServerProcessing = s.serverDone.Sub(s.serverStart)
			s.StartTransfer = s.serverDone.Sub(s.dnsStart)
			s.transferStart = s.serverDone
		},
	})
}
