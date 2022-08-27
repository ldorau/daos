//
// (C) Copyright 2022 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package storage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/daos-stack/daos/src/control/common/test"
)

func Test_NvmeDevState(t *testing.T) {
	for name, tc := range map[string]struct {
		state        NvmeDevState
		expIsInUse   bool
		expIsPlugged bool
		expIsFaulty  bool
		expIsInvalid bool
		expStr       string
	}{
		"new state": {
			state:        NvmeStateNew,
			expIsPlugged: true,
			expStr:       "NEW",
		},
		"normal state": {
			state:        NvmeStateNormal,
			expIsPlugged: true,
			expIsInUse:   true,
			expStr:       "NORMAL",
		},
		"faulty state": {
			state:        NvmeStateFaulty,
			expIsPlugged: true,
			expIsInUse:   true,
			expIsFaulty:  true,
			expStr:       "EVICTED",
		},
		// Although this should not be returned, worth testing just in case.
		"plugged and faulty flags; not in use": {
			state:        NvmeFlagPlugged | NvmeFlagFaulty,
			expIsPlugged: true,
			expIsFaulty:  true,
			expStr:       "EVICTED",
		},
		"unplugged; no flags": {
			state:  0,
			expStr: "UNPLUGGED",
		},
		"invalid flag": {
			state:        NvmeFlagInvalid,
			expIsInvalid: true,
			expStr:       "INVALID",
		},
	} {
		t.Run(name, func(t *testing.T) {
			test.AssertEqual(t, tc.expIsInUse, tc.state&NvmeFlagInUse != 0,
				"unexpected IsInUse check")
			test.AssertEqual(t, tc.expIsFaulty, tc.state&NvmeFlagFaulty != 0,
				"unexpected IsFaulty check")
			test.AssertEqual(t, tc.expIsPlugged, tc.state&NvmeFlagPlugged != 0,
				"unexpected IsPlugged check")
			test.AssertEqual(t, tc.expIsInvalid, tc.state.IsInvalid(),
				"unexpected IsInvalid() result")

			test.AssertEqual(t, tc.expStr, tc.state.String(),
				"unexpected status string")

			stateNew := NvmeDevStateFromString(tc.state.String())

			test.AssertEqual(t, tc.state.String(), stateNew.String(),
				fmt.Sprintf("expected string %s to yield state %s",
					tc.state.String(), stateNew.String()))
		})
	}
}

func Test_VmdLedState(t *testing.T) {
	for name, tc := range map[string]struct {
		state  VmdLedState
		expStr string
	}{
		"normal state": {
			state:  LedStateNormal,
			expStr: "OFF",
		},
		"faulty state": {
			state:  LedStateFaulty,
			expStr: "ON",
		},
		"identify state": {
			state:  LedStateIdentify,
			expStr: "QUICK-BLINK",
		},
		"unknown and unsupported state": {
			state:  LedStateUnknown,
			expStr: "NA",
		},
	} {
		t.Run(name, func(t *testing.T) {
			test.AssertEqual(t, tc.expStr, tc.state.String(),
				"unexpected status string")
		})
	}
}

func Test_filterBdevScanResponse(t *testing.T) {
	const (
		vmdAddr1         = "0000:5d:05.5"
		vmdBackingAddr1a = "5d0505:01:00.0"
		vmdBackingAddr1b = "5d0505:03:00.0"
		vmdAddr2         = "0000:7d:05.5"
		vmdBackingAddr2a = "7d0505:01:00.0"
		vmdBackingAddr2b = "7d0505:03:00.0"
	)
	ctrlrsFromPCIAddrs := func(addrs ...string) (ncs NvmeControllers) {
		for _, addr := range addrs {
			ncs = append(ncs, &NvmeController{PciAddr: addr})
		}
		return
	}

	for name, tc := range map[string]struct {
		addrs    []string
		scanResp *BdevScanResponse
		expAddrs []string
		expErr   error
	}{
		"two vmd endpoints; one filtered out": {
			addrs: []string{vmdAddr2},
			scanResp: &BdevScanResponse{
				Controllers: ctrlrsFromPCIAddrs(vmdBackingAddr1a, vmdBackingAddr1b,
					vmdBackingAddr2a, vmdBackingAddr2b),
			},
			expAddrs: []string{vmdBackingAddr2a, vmdBackingAddr2b},
		},
		"two ssds; one filtered out": {
			addrs: []string{"0000:81:00.0"},
			scanResp: &BdevScanResponse{
				Controllers: ctrlrsFromPCIAddrs("0000:81:00.0", "0000:de:00.0"),
			},
			expAddrs: []string{"0000:81:00.0"},
		},
		"two aio kdev paths; both filtered out": {
			addrs: []string{"/dev/sda"},
			scanResp: &BdevScanResponse{
				Controllers: ctrlrsFromPCIAddrs("/dev/sda", "/dev/sdb"),
			},
			expAddrs: []string{},
		},
		"bad address; filtered out": {
			addrs: []string{"0000:81:00.0"},
			scanResp: &BdevScanResponse{
				Controllers: ctrlrsFromPCIAddrs("0000:81.00.0"),
			},
			expAddrs: []string{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			bdl := new(BdevDeviceList)
			if err := bdl.fromStrings(tc.addrs); err != nil {
				t.Fatal(err)
			}
			gotErr := filterBdevScanResponse(bdl, tc.scanResp)
			test.CmpErr(t, tc.expErr, gotErr)
			if gotErr != nil {
				return
			}

			expAddrStr := strings.Join(tc.expAddrs, ", ")
			if diff := cmp.Diff(expAddrStr, tc.scanResp.Controllers.String()); diff != "" {
				t.Fatalf("unexpected output addresses (-want, +got):\n%s\n", diff)
			}
		})
	}
}
