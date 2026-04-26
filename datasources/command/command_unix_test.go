package command

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"
	"github.com/duakc/mt/sh"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	var (
		ip1 = mt.Must(netip.ParseAddr("1.1.1.1"))
		ip2 = mt.Must(netip.ParseAddr("2001:4860:4860::8888"))
		ip3 = mt.Must(netip.ParseAddr("1.0.0.1"))
	)

	cases := []struct {
		cmd   string
		shCmd *sh.Cmd

		exitCode int
		ip       []netip.Addr
		err      error
	}{
		{cmd: "echo 1.1.1.1", ip: []netip.Addr{ip1}},
		{cmd: "echo 1.1.1.1 && exit 1;", ip: []netip.Addr{ip1}, exitCode: 1},
		{cmd: "echo 2001:4860:4860::8888", ip: []netip.Addr{ip2}},
		{cmd: `echo -e "1.1.1.1\n2001:4860:4860::8888\r\n1.0.0.1"`, ip: []netip.Addr{
			ip1, ip2, ip3,
		}},
		{cmd: `echo -e "1.1.1.1 2001:4860:4860::8888\n1.0.0.1\n\n\r\n\n\r\t\n"`, ip: []netip.Addr{
			ip1, ip2, ip3,
		}},
		{cmd: `echo -e "\"1.1.1.1\" 2001:4860:4860::8888\n\"1.0.0.1\"\n\n\r\n\n\r\t\n"`, ip: []netip.Addr{
			ip1, ip2, ip3,
		}},
	}

	for _, cc := range cases {
		const dur = 1 * time.Second
		timeout, cancel := context.WithTimeout(context.Background(), dur)
		cmd := cc.shCmd
		if cmd == nil {
			cmd = sh.New().Deattach()
		}
		ip, err := runCommand(timeout, zaplog.NOP, cmd, cc.cmd, cc.exitCode)
		cancel()

		if cc.err == nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, cc.err, err)
		}
		assert.Equal(t, cc.ip, ip)
	}
}
