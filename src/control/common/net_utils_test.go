//
// (C) Copyright 2020-2022 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package common

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/daos-stack/daos/src/control/common/test"
	. "github.com/daos-stack/daos/src/control/common/test"
)

func TestUtils_HasPort(t *testing.T) {
	for name, tc := range map[string]struct {
		addr   string
		expRes bool
	}{
		"host has port": {"localhost:10001", true},
		"host no port":  {"localhost", false},
		"ip has port":   {"192.168.1.1:10001", true},
		"ip no port":    {"192.168.1.1", false},
	} {
		t.Run(name, func(t *testing.T) {
			AssertEqual(t, tc.expRes, HasPort(tc.addr), name)
		})
	}
}

func TestUtils_SplitPort(t *testing.T) {
	for name, tc := range map[string]struct {
		addr      string
		dPort     int
		expHost   string
		expPort   string
		expErrMsg string
	}{
		"host has port": {"localhost:10001", 10000, "localhost", "10001", ""},
		"host no port":  {"localhost", 10000, "localhost", "10000", ""},
		"ip has port":   {"192.168.1.1:10001", 10000, "192.168.1.1", "10001", ""},
		"ip no port":    {"192.168.1.1", 10000, "192.168.1.1", "10000", ""},
		"empty port":    {"192.168.1.1:", 10000, "", "", "invalid port \"\""},
		"bad port":      {"192.168.1.1:abc", 10000, "", "", "invalid port \"abc\""},
		"bad address": {"192.168.1.1:10001:", 10000, "", "",
			"address 192.168.1.1:10001:: too many colons in address"},
	} {
		t.Run(name, func(t *testing.T) {
			h, p, err := SplitPort(tc.addr, tc.dPort)
			ExpectError(t, err, tc.expErrMsg, name)
			if tc.expErrMsg != "" {
				return
			}

			AssertEqual(t, tc.expHost, h, name)
			AssertEqual(t, tc.expPort, p, name)
		})
	}
}

func TestCommon_CmpTCPAddr(t *testing.T) {
	testA := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
	testB := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2)}
	testC := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1}
	testD := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1, Zone: "quack"}

	for name, tc := range map[string]struct {
		a      *net.TCPAddr
		b      *net.TCPAddr
		expRes bool
	}{
		"both nil": {
			a:      nil,
			b:      nil,
			expRes: true,
		},
		"a nil": {
			a: nil,
			b: testB,
		},
		"b nil": {
			a: testA,
			b: nil,
		},
		"same": {
			a:      testA,
			b:      testA,
			expRes: true,
		},
		"different IP": {
			a: testA,
			b: testB,
		},
		"different port": {
			a: testA,
			b: testC,
		},
		"different zone": {
			a: testA,
			b: testD,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.expRes, CmpTCPAddr(tc.a, tc.b)); diff != "" {
				t.Fatalf("unexpected result (-want, +got):\n%s\n", diff)
			}
		})
	}
}

func TestCommon_IsLocalAddr(t *testing.T) {
	local := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
	remote := &net.TCPAddr{IP: net.IPv4(127, 127, 127, 127)}

	for name, tc := range map[string]struct {
		a      *net.TCPAddr
		expRes bool
	}{
		"nil": {},
		"local": {
			a:      local,
			expRes: true,
		},
		"remote": {
			a: remote,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.expRes, IsLocalAddr(tc.a)); diff != "" {
				t.Fatalf("unexpected result (-want, +got):\n%s\n", diff)
			}
		})
	}
}

func TestCommon_ParseHostList(t *testing.T) {
	testPort := 12345

	mockHostList := func(hosts ...string) []string {
		return hosts
	}

	for name, tc := range map[string]struct {
		in     []string
		expOut []string
		expErr error
	}{
		"nil in, nil out": {},
		"empty in, nil out": {
			in: []string{},
		},
		"host with too many :": {
			in:     mockHostList("foo::10"),
			expErr: errors.New("invalid host"),
		},
		"host with non-numeric port": {
			in:     mockHostList("foo:bar"),
			expErr: errors.New("invalid host"),
		},
		"host with just a port": {
			in:     mockHostList(":42"),
			expErr: errors.New("invalid host"),
		},
		"should append missing port": {
			in:     mockHostList("foo"),
			expOut: mockHostList(fmt.Sprintf("foo:%d", testPort)),
		},
		"should append missing port (multiple)": {
			in: mockHostList("foo", "bar:4242", "baz"),
			expOut: mockHostList(
				"bar:4242",
				fmt.Sprintf("baz:%d", testPort),
				fmt.Sprintf("foo:%d", testPort),
			),
		},
		"should append missing port (ranges)": {
			in: mockHostList("foo-[1-4]", "bar[2-4]", "baz[8-9]:4242"),
			expOut: mockHostList(
				fmt.Sprintf("bar2:%d", testPort),
				fmt.Sprintf("bar3:%d", testPort),
				fmt.Sprintf("bar4:%d", testPort),
				"baz8:4242",
				"baz9:4242",
				fmt.Sprintf("foo-1:%d", testPort),
				fmt.Sprintf("foo-2:%d", testPort),
				fmt.Sprintf("foo-3:%d", testPort),
				fmt.Sprintf("foo-4:%d", testPort),
			),
		},
	} {
		t.Run(name, func(t *testing.T) {
			gotOut, gotErr := ParseHostList(tc.in, testPort)
			CmpErr(t, tc.expErr, gotErr)
			if tc.expErr != nil {
				return
			}
			if diff := cmp.Diff(tc.expOut, gotOut); diff != "" {
				t.Fatalf("unexpected output (-want, +got):\n%s\n", diff)
			}
		})
	}

}

func TestCommon_CheckIPOnNetwork(t *testing.T) {
	for name, tc := range map[string]struct {
		addr   net.IP
		ipNet  *net.IPNet
		expErr error
	}{
		"nil IPnet": {
			expErr: errors.New("nil"),
		},
		"ipv6 addr with ipv4 net": {
			addr: net.ParseIP("::1"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("127.0.0.1"),
				Mask: net.IPv4Mask(0xff, 0xff, 0, 0),
			},
			expErr: errors.New("not comparable IP versions"),
		},
		"ipv4 addr with ipv6 net": {
			addr: net.ParseIP("127.0.0.1"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("::1"),
				Mask: net.CIDRMask(32, 128),
			},
			expErr: errors.New("not comparable IP versions"),
		},
		"ipv6 addr not on net": {
			addr: net.ParseIP("dead::1234:5678:9abc:deff"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("deaf::1234:5678:9abc:deff"),
				Mask: net.CIDRMask(32, 128),
			},
			expErr: errors.New("not on network"),
		},
		"ipv4 addr not on net": {
			addr: net.ParseIP("10.108.36.4"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("10.109.36.4"),
				Mask: net.IPv4Mask(0xff, 0xff, 0, 0),
			},
			expErr: errors.New("not on network"),
		},
		"ipv6 addr on net": {
			addr: net.ParseIP("dead::1234:5678:9abc:deff"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("dead::ffed:cba9:8765:4321"),
				Mask: net.CIDRMask(32, 128),
			},
		},
		"ipv4 addr on net": {
			addr: net.ParseIP("10.108.36.4"),
			ipNet: &net.IPNet{
				IP:   net.ParseIP("10.108.200.10"),
				Mask: net.IPv4Mask(0xff, 0xff, 0, 0),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := CheckIPAddrOnNetwork(tc.addr, tc.ipNet)

			test.CmpErr(t, tc.expErr, err)
		})
	}
}

func TestCommon_SameIPVersion(t *testing.T) {
	for name, tc := range map[string]struct {
		ip1       net.IP
		ip2       net.IP
		expResult bool
	}{
		"both empty": {
			expResult: true,
		},
		"empty vs ipv4": {
			ip1: net.ParseIP("127.0.0.1"),
		},
		"empty vs ipv6": {
			ip2: net.ParseIP("::1"),
		},
		"both ipv4": {
			ip1:       net.ParseIP("127.0.0.1"),
			ip2:       net.ParseIP("10.232.128.15"),
			expResult: true,
		},
		"both ipv6": {
			ip1:       net.ParseIP("deaf::1234:5678:9abc:deff"),
			ip2:       net.ParseIP("fe80::1234:5678:9abc:deff"),
			expResult: true,
		},
		"ipv4 vs ipv6": {
			ip1: net.ParseIP("127.0.0.1"),
			ip2: net.ParseIP("fe80::1234:5678:9abc:deff"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			test.AssertEqual(t, tc.expResult, SameIPVersion(tc.ip1, tc.ip2), "")
		})
	}
}
