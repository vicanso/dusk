// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dusk

import (
	"crypto/tls"
	"net/http/httptrace"
	"sync"
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
	// HTTPTrace http trace
	HTTPTrace struct {
		// 因为timeout的设置有可能导致 trace 读写并存，因此需要锁
		sync.RWMutex
		Host           string        `json:"host,omitempty"`
		Addrs          []string      `json:"addrs,omitempty"`
		Network        string        `json:"network,omitempty"`
		Addr           string        `json:"addr,omitempty"`
		Reused         bool          `json:"reused,omitempty"`
		WasIdle        bool          `json:"wasIdle,omitempty"`
		IdleTime       time.Duration `json:"idleTime,omitempty"`
		Protocol       string        `json:"protocol,omitempty"`
		TLSVersion     string        `json:"tlsVersion,omitempty"`
		TLSResume      bool          `json:"tlsResume,omitempty"`
		TLSCipherSuite string        `json:"tlsCipherSuite,omitempty"`

		Start                time.Time `json:"start,omitempty"`
		DNSStart             time.Time `json:"dnsStart,omitempty"`
		DNSDone              time.Time `json:"dnsDone,omitempty"`
		ConnectStart         time.Time `json:"connectStart,omitempty"`
		ConnectDone          time.Time `json:"connectDone,omitempty"`
		GotConnect           time.Time `json:"gotConnect,omitempty"`
		GotFirstResponseByte time.Time `json:"gotFirstResponseByte,omitempty"`
		TLSHandshakeStart    time.Time `json:"tlsHandshakeStart,omitempty"`
		TLSHandshakeDone     time.Time `json:"tlsHandshakeDone,omitempty"`
		Done                 time.Time `json:"done,omitempty"`
	}
)

var (
	versions     map[uint16]string
	cipherSuites map[uint16]string
)

const (
	unknown = "unknown"
)

func init() {
	versions = map[uint16]string{
		tls.VersionSSL30: "ssl3.0",
		tls.VersionTLS10: "tls1.0",
		tls.VersionTLS11: "tls1.1",
		tls.VersionTLS12: "tls1.2",
		versionTLS13:     "tls1.3",
	}
	cipherSuites = map[uint16]string{
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "ECDHE_RSA_WITH_CHACHA20_POLY1305",
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "ECDHE_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "ECDHE_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256:         "RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_RC4_128_SHA:                "RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "ECDHE_RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "ECDHE_ECDSA_WITH_RC4_128_SHA",
	}
}

func convertTLSVersion(version uint16) string {
	v := versions[version]
	if v == "" {
		v = unknown
	}
	return v
}

func convertCipherSuite(cipherSuite uint16) string {
	v := cipherSuites[cipherSuite]
	if v == "" {
		v = unknown
	}
	return v
}

// Finish http trace finish
func (ht *HTTPTrace) Finish() {
	ht.Lock()
	defer ht.Unlock()
	ht.Done = time.Now()
}

// Stats get the stats of time line
func (ht *HTTPTrace) Stats() (stats *HTTPTimelineStats) {
	stats = &HTTPTimelineStats{}
	ht.RLock()
	defer ht.RUnlock()
	if !ht.DNSStart.IsZero() && !ht.DNSDone.IsZero() {
		stats.DNSLookup = ht.DNSDone.Sub(ht.DNSStart)
	}
	if !ht.ConnectStart.IsZero() && !ht.ConnectDone.IsZero() {
		stats.TCPConnection = ht.ConnectDone.Sub(ht.ConnectStart)
	}
	if !ht.TLSHandshakeStart.IsZero() && !ht.TLSHandshakeDone.IsZero() {
		stats.TLSHandshake = ht.TLSHandshakeDone.Sub(ht.TLSHandshakeStart)
	}

	if !ht.GotConnect.IsZero() && !ht.GotFirstResponseByte.IsZero() {
		stats.ServerProcessing = ht.GotFirstResponseByte.Sub(ht.GotConnect)
	}
	if ht.Done.IsZero() {
		ht.Done = time.Now()
	}
	if !ht.GotFirstResponseByte.IsZero() {
		stats.ContentTransfer = ht.Done.Sub(ht.GotFirstResponseByte)
	}
	stats.Total = ht.Done.Sub(ht.Start)
	return
}

// NewClientTrace http client trace
func NewClientTrace() (trace *httptrace.ClientTrace, ht *HTTPTrace) {
	ht = &HTTPTrace{
		Start: time.Now(),
	}
	trace = &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			ht.Lock()
			defer ht.Unlock()
			ht.Host = info.Host
			ht.DNSStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			ht.Lock()
			defer ht.Unlock()
			ht.Addrs = make([]string, len(info.Addrs))
			for index, addr := range info.Addrs {
				ht.Addrs[index] = addr.String()
			}
			ht.DNSDone = time.Now()
		},
		ConnectStart: func(network, addr string) {
			ht.Lock()
			defer ht.Unlock()
			ht.Network = network
			ht.Addr = addr
			ht.ConnectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			ht.Lock()
			defer ht.Unlock()
			ht.ConnectDone = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			ht.Lock()
			defer ht.Unlock()
			ht.Reused = info.Reused
			ht.WasIdle = info.WasIdle
			ht.IdleTime = info.IdleTime

			ht.GotConnect = time.Now()
		},
		GotFirstResponseByte: func() {
			ht.Lock()
			defer ht.Unlock()
			ht.GotFirstResponseByte = time.Now()
		},
		TLSHandshakeStart: func() {
			ht.Lock()
			defer ht.Unlock()
			ht.TLSHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(info tls.ConnectionState, _ error) {
			ht.Lock()
			defer ht.Unlock()
			ht.TLSVersion = convertTLSVersion(info.Version)
			ht.TLSResume = info.DidResume
			ht.TLSCipherSuite = convertCipherSuite(info.CipherSuite)
			ht.Protocol = info.NegotiatedProtocol

			ht.TLSHandshakeDone = time.Now()
		},
	}
	return
}
