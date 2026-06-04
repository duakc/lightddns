package ipserver

import (
	"net"
	"net/netip"
	"testing"
)

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
