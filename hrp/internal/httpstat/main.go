// Package httpstat traces HTTP latency infomation (DNSLookup, TCP Connection and so on) on any golang HTTP request.
// It uses `httptrace` package.
// Inspired by https://github.com/tcnksm/go-httpstat and https://github.com/davecheney/httpstat
package httpstat

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

const (
	httpsTemplate = "\n" +
		`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
		`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
		`            |                |               |                   |                  |` + "\n" +
		`   namelookup:%s      |               |                   |                  |` + "\n" +
		`                       connect:%s     |                   |                  |` + "\n" +
		`                                   pretransfer:%s         |                  |` + "\n" +
		`                                                     starttransfer:%s        |` + "\n" +
		`                                                                                total:%s` + "\n\n"

	httpTemplate = "\n" +
		`   DNS Lookup   TCP Connection   Server Processing   Content Transfer` + "\n" +
		`[ %s  |     %s  |        %s  |       %s  ]` + "\n" +
		`             |                |                   |                  |` + "\n" +
		`    namelookup:%s      |                   |                  |` + "\n" +
		`                        connect:%s         |                  |` + "\n" +
		`                                      starttransfer:%s        |` + "\n" +
		`                                                                 total:%s` + "\n\n"
)

func fmta(d time.Duration) string {
	return color.BlueString("%7dms", int(d.Milliseconds()))
}

func fmtb(d time.Duration) string {
	return color.MagentaString("%-9s", strconv.Itoa(int(d.Milliseconds()))+"ms")
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}

func colorize(s string) string {
	v := strings.Split(s, "\n")
	v[0] = grayscale(16)(v[0])
	return strings.Join(v, "\n")
}

func printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

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

func (s *Stat) assertSchemaHTTP() bool {
	// HTTP other than HTTPS
	return s.dnsStart.IsZero() && s.tcpStart.IsZero()
}

// Durations returns all durations and timelines of request latencies
func (s *Stat) Durations() map[string]int64 {
	return map[string]int64{
		"DNSLookup":        s.DNSLookup.Milliseconds(),
		"TCPConnection":    s.TCPConnection.Milliseconds(),
		"TLSHandshake":     s.TLSHandshake.Milliseconds(),
		"ServerProcessing": s.ServerProcessing.Milliseconds(),
		"ContentTransfer":  s.ContentTransfer.Milliseconds(),
		"NameLookup":       s.NameLookup.Milliseconds(),
		"Connect":          s.Connect.Milliseconds(),
		"Pretransfer":      s.Pretransfer.Milliseconds(),
		"StartTransfer":    s.StartTransfer.Milliseconds(),
		"Total":            s.Total.Milliseconds(),
	}
}

func (s *Stat) Print() {
	if s.assertSchemaHTTP() {
		// http
		printf(colorize(httpTemplate),
			fmta(s.DNSLookup),        // dns lookup
			fmta(s.TCPConnection),    // tcp connection
			fmta(s.ServerProcessing), // server processing
			fmta(s.ContentTransfer),  // content transfer
			fmtb(s.NameLookup),       // namelookup
			fmtb(s.Connect),          // connect
			fmtb(s.StartTransfer),    // starttransfer
			fmtb(s.Total),            // total
		)
	} else {
		// https
		printf(colorize(httpsTemplate),
			fmta(s.DNSLookup),        // dns lookup
			fmta(s.TCPConnection),    // tcp connection
			fmta(s.TLSHandshake),     // tls handshake
			fmta(s.ServerProcessing), // server processing
			fmta(s.ContentTransfer),  // content transfer
			fmtb(s.NameLookup),       // namelookup
			fmtb(s.Connect),          // connect
			fmtb(s.Pretransfer),      // pretransfer
			fmtb(s.StartTransfer),    // starttransfer
			fmtb(s.Total),            // total
		)
	}
	log.Info().
		Interface("httpstat(ms)", s.Durations()).
		Msg("HTTP latency statistics")
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
			if s.assertSchemaHTTP() {
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
