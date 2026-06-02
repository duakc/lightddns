package ipserver

import (
	"net"
	"net/netip"
	"testing"
)

// goos: darwin
// goarch: arm64
// pkg: github.com/duakc/lightddns/services/ipserver
// cpu: Apple M5
// BenchmarkParseAddr
// BenchmarkParseAddr/SplitFirst
// BenchmarkParseAddr/SplitFirst-10         	90424994	        12.72 ns/op
// BenchmarkParseAddr/ParseAddrPort
// BenchmarkParseAddr/ParseAddrPort-10      	48085351	        24.51 ns/op
// BenchmarkParseAddr/ParseAddrPortNoHost
// BenchmarkParseAddr/ParseAddrPortNoHost-10         	78858734	        14.87 ns/op
func BenchmarkParseAddr(b *testing.B) {
	testData := "1.1.1.1:443"

	b.Run("SplitFirst", func(b *testing.B) {
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			var host string
			var err error
			if host, _, err = net.SplitHostPort(testData); err != nil {
				b.Fatal(err)
			}

			if _, err = netip.ParseAddr(host); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
	})

	b.Run("ParseAddrPort", func(b *testing.B) {
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			var host string
			var err error
			addrPort, err := netip.ParseAddrPort(testData)
			if err != nil {
				b.Fatal(err)
			}
			host = addrPort.Addr().String()
			_ = host
		}
	})

	b.Run("ParseAddrPortNoHost", func(b *testing.B) {
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			if _, err := netip.ParseAddrPort(testData); err != nil {
				b.Fatal(err)
			}
		}
	})
}
