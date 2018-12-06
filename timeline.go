package dusk

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

type (
	// HTTPTimelineStats http timeline stats
	HTTPTimelineStats struct {
		DNSLookup        time.Duration `json:"dnsLookup,omitempty"`
		TCPConnection    time.Duration `json:"tcpConnection,omitempty"`
		TLSHandshake     time.Duration `json:"tlsHandshake,omitempty"`
		ServerProcessing time.Duration `json:"serverProcessing,omitempty"`
		ContentTransfer  time.Duration `json:"contentTransfer,omitempty"`
		Total            time.Duration `json:"total,omitempty"`
	}
	// HTTPTimeline http timing
	HTTPTimeline struct {
		Start                time.Time
		DNSStart             time.Time
		DNSDone              time.Time
		ConnectStart         time.Time
		ConnectDone          time.Time
		GotConnect           time.Time
		GotFirstResponseByte time.Time
		TLSHandshakeStart    time.Time
		TLSHandshakeDone     time.Time
		Done                 time.Time
	}
)

// Stats get the stats of time line
func (tl *HTTPTimeline) Stats() (stats *HTTPTimelineStats) {
	stats = &HTTPTimelineStats{}
	if !tl.DNSStart.IsZero() {
		stats.DNSLookup = tl.DNSDone.Sub(tl.DNSStart)
	}
	if !tl.ConnectStart.IsZero() {
		stats.TCPConnection = tl.ConnectDone.Sub(tl.ConnectStart)
	}
	if !tl.TLSHandshakeStart.IsZero() {
		stats.TLSHandshake = tl.TLSHandshakeDone.Sub(tl.TLSHandshakeStart)
	}

	if !tl.GotConnect.IsZero() {
		stats.ServerProcessing = tl.GotFirstResponseByte.Sub(tl.GotConnect)
	}
	if tl.Done.IsZero() {
		tl.Done = time.Now()
	}
	if !tl.GotFirstResponseByte.IsZero() {
		stats.ContentTransfer = tl.Done.Sub(tl.GotFirstResponseByte)
	}
	stats.Total = tl.Done.Sub(tl.Start)
	return
}

// NewClientTrace http client trace
func NewClientTrace() (trace *httptrace.ClientTrace, tl *HTTPTimeline) {
	tl = &HTTPTimeline{
		Start: time.Now(),
	}
	trace = &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			tl.DNSStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			tl.DNSDone = time.Now()
		},
		ConnectStart: func(_, _ string) {
			tl.ConnectStart = time.Now()
		},
		ConnectDone: func(net, addr string, err error) {
			tl.ConnectDone = time.Now()
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			tl.GotConnect = time.Now()
		},
		GotFirstResponseByte: func() {
			tl.GotFirstResponseByte = time.Now()
		},
		TLSHandshakeStart: func() {
			tl.TLSHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			tl.TLSHandshakeDone = time.Now()
		},
	}
	return
}
