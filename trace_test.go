package dusk

import (
	"crypto/tls"
	"net"
	"net/http/httptrace"
	"testing"
	"time"
)

func TestConvertTLSVersion(t *testing.T) {
	if convertTLSVersion(tls.VersionSSL30) != "ssl3.0" {
		t.Fatalf("convert tls version fail")
	}
	if convertTLSVersion(1) != "unknown" {
		t.Fatalf("convert tls version fail")
	}
}

func TestConvertCipherSuite(t *testing.T) {
	if convertCipherSuite(tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA) != "ECDHE_RSA_WITH_AES_128_CBC_SHA" {
		t.Fatalf("convert cipher suite fail")
	}
	if convertCipherSuite(1) != "unknown" {
		t.Fatalf("convert cipher suite fail")
	}
}

func TestTrace(t *testing.T) {
	trace, ht := NewClientTrace()

	trace.DNSStart(httptrace.DNSStartInfo{
		Host: "aslant.site",
	})
	time.Sleep(time.Millisecond)

	addrs := make([]net.IPAddr, 0)
	addrs = append(addrs, net.IPAddr{
		IP: net.IPv4(1, 1, 1, 1),
	})
	trace.DNSDone(httptrace.DNSDoneInfo{
		Addrs: addrs,
	})
	time.Sleep(time.Millisecond)

	trace.ConnectStart("tcp", "1.1.1.1")
	time.Sleep(time.Millisecond)

	trace.ConnectDone("", "", nil)
	time.Sleep(time.Millisecond)

	trace.TLSHandshakeStart()
	time.Sleep(time.Millisecond)

	trace.TLSHandshakeDone(tls.ConnectionState{}, nil)
	time.Sleep(time.Millisecond)

	trace.GotConn(httptrace.GotConnInfo{
		Reused:  true,
		WasIdle: true,
	})
	time.Sleep(time.Millisecond)

	trace.GotFirstResponseByte()
	time.Sleep(time.Millisecond)

	stats := ht.Stats()
	if stats.DNSLookup == 0 ||
		stats.TCPConnection == 0 ||
		stats.TLSHandshake == 0 ||
		stats.ServerProcessing == 0 ||
		stats.ContentTransfer == 0 ||
		stats.Total == 0 {
		t.Fatalf("get http stats fail")
	}
}
