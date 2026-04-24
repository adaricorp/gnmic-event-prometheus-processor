package main

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/openconfig/gnmic/pkg/api/types"
	"github.com/openconfig/gnmic/pkg/formatters"
	"github.com/stretchr/testify/assert"
)

var (
	apTags = &[]string{
		"ap-name",
		"ap-mac",
		"site-tag",
		"location-name",
	}
	bypassApTags = &[]string{
		"ap-global-oper-data/ap-join-stats",
		"access-point-oper-data/capwap-pkts",
	}
	tagRenames = map[string]map[string]string{
		"joined-aps": {
			"hostname": "ap-name",
			"mac":      "eth-mac",
		},
		"ap-global-oper-data/ap-join-stats": {
			"wtp-mac":         "ap-mac",
			"ap-ethernet-mac": "eth-mac",
		},
		"access-point-oper-data/capwap-data": {
			"wtp-serial-num": "serial-number",
		},
		"access-point-oper-data/capwap-pkts": {
			"wtp-mac": "ap-mac",
		},
		"access-point-oper-data/radio-oper-data": {
			"curr-freq":           "channel",
			"current-band-id":     "band-id",
			"current-active-band": "band",
		},
		"access-point-oper-data": {
			"radio-slot-id": "slot-id",
		},
		"rrm-oper-data": {
			"chan":          "channel",
			"radio-slot-id": "slot-id",
		},
		"client-oper-data": {
			"ms-mac-address": "client-mac",
			"ms-ap-slot-id":  "slot-id",
		},
		"client-oper-data/policy-data": {
			"mac": "client-mac",
		},
		"client-oper-data/dot11-oper-data": {
			"ap-mac-address":  "eth-mac",
			"current-channel": "channel",
			"ms-wlan-id":      "wlan-id",
			"vap-ssid":        "ssid",
		},
	}
	valueTags = map[string][]string{
		"access-point-oper-data/radio-oper-data": {
			"current-band-id",
		},
		"access-point-oper-data/capwap-data": {
			"numeric-id",
		},
		"access-point-oper-data/ethernet-if-stats": {
			"if-name",
		},
		"client-oper-data": {
			"ms-ap-slot-id",
		},
		"client-oper-data/dot11-oper-data": {
			"ap-mac-address",
			"ms-wlan-id",
			"vap-ssid",
		},
		"ap-global-oper-data/ap-join-stats": {
			"ap-name",
		},
		"rrm-oper-data/rrm-measurement/noise": {
			"chan",
		},
	}
	enumMappings = map[string]map[string]map[string]int{
		"access-point-oper-data/ethernet-if-stats": {
			"oper-status": {
				"oper-state-down": 0,
				"oper-state-up":   1,
				"oper-state-na":   2,
			},
		},
		"joined-aps": {
			"opstate": {
				"DOWN":      0,
				"UP":        1,
				"UPGRADING": 2,
			},
		},
	}
	integerInfoMetrics = map[string][]string{
		"access-point-oper-data/radio-oper-data": {
			"curr-freq",
		},
		"client-oper-data/dot11-oper-data": {
			"current-channel",
		},
	}
	valueAndInfoMetrics = map[string][]string{
		"access-point-oper-data/radio-oper-data": {
			"curr-freq",
		},
		"access-point-oper-data/ethernet-if-stats": {
			"oper-status",
		},
		"client-oper-data/policy-data": {
			"res-vlan-id",
		},
		"client-oper-data/dot11-oper-data": {
			"current-channel",
		},
	}
	mixedInfoMetrics = &[]string{
		"device-hardware-data/device-hardware/device-inventory",
		"access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver",
		"geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center",
	}
	valueBlocklist = &[]string{
		"ap-cfg-data/location-entries/location-entry/location-name",
		"device-hardware-data/device-hardware/device-inventory/hw-type",
	}
)

type testCall struct {
	name     string
	input    []*formatters.EventMsg
	expected []*formatters.EventMsg
}

func allYangPathCombinations(yangPath string, descend bool) []string {
	yangPath = stripYangPathOrigin(yangPath)

	yangPaths := []string{}

	currDir := yangPath
	for {
		for _, variation := range []string{
			"/" + stripYangPathOrigin(currDir) + "/",
			"/" + stripYangPathOrigin(currDir),
			stripYangPathOrigin(currDir) + "/",
			stripYangPathOrigin(currDir),
			"/" + stripYangPathNamespace(currDir) + "/",
			"/" + stripYangPathNamespace(currDir),
			stripYangPathNamespace(currDir) + "/",
			stripYangPathNamespace(currDir),
		} {
			if !slices.Contains(yangPaths, variation) {
				yangPaths = append(yangPaths, variation)
			}
		}

		if !descend {
			break
		}

		currDir = filepath.Dir(currDir)

		if currDir == "." || currDir == "/" {
			break
		}
	}

	return yangPaths
}

func TestApply(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		calls []testCall
	}

	testCases := []testCase{
		{
			name: "wlc",
			calls: []testCall{
				{
					name: "system/config/hostname",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
				{
					name: "device-hardware-data/device-hardware/device-inventory",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "wlc",
								"device-inventory_hw-type": "hw-type-chassis",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-type": "hw-type-chassis",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 0,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/part-number": "C9800-CL-K9",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "wlc",
								"device-inventory_hw-type": "hw-type-pim",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-type": "hw-type-pim",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-pim",
								"device-inventory_hw-dev-index": "1",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 1,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-pim",
								"device-inventory_hw-dev-index": "1",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-type":           "hw-type-chassis",
								"hw-dev-index":      "0",
								"version":           "V00",
								"part-number":       "C9800-CL-K9",
								"field-replaceable": "false",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-type":           "hw-type-pim",
								"hw-dev-index":      "1",
								"version":           "V00",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "access points",
			calls: []testCall{
				{
					name: "access-point-oper-data/oper-data/ap-sys-stats/cpu-usage fail",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 2,
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map ap01",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/wtp-mac": "01:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/eth-mac": "00:00:00:00:00:01",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map ap02",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/wtp-mac": "01:00:00:00:00:02",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/eth-mac": "00:00:00:00:00:02",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "ap-cfg-data/ap-tags/ap-tag/site-tag",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-tag_ap-mac":     "00:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/ap-tags/ap-tag/ap-mac": "00:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-tag_ap-mac":     "00:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/ap-tags/ap-tag/site-tag": "Site1",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-tag_ap-mac":     "00:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/ap-tags/ap-tag/ap-mac": "00:00:00:00:00:02",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-tag_ap-mac":     "00:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/ap-tags/ap-tag/site-tag": "Site2",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "ap-cfg-data/location-entries/location-entry",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                       "wlc.example.org",
								"subscription-name":            "ap",
								"location-entry_location-name": "Location1",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/location-entries/location-entry/location-name": "Location1",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                       "wlc.example.org",
								"subscription-name":            "ap",
								"associated-ap_ap-mac":         "00:00:00:00:00:01",
								"location-entry_location-name": "Location1",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/location-entries/location-entry/associated-aps/associated-ap/ap-mac": "00:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                       "wlc.example.org",
								"subscription-name":            "ap",
								"location-entry_location-name": "Location2",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/location-entries/location-entry/location-name": "Location2",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                       "wlc.example.org",
								"subscription-name":            "ap",
								"associated-ap_ap-mac":         "00:00:00:00:00:02",
								"location-entry_location-name": "Location2",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-cfg:ap-cfg-data/location-entries/location-entry/associated-aps/associated-ap/ap-mac": "00:00:00:00:00:02",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "joined-aps",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/mac": "00:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/hostname": "ap01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/opstate": "UP",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/uptime": 1295277,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/serial": "ABCD1234",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"joined-ap_hostname": "ap01",
							},
							Values: map[string]any{
								"/joined-aps/joined-ap/state/model": "C9105AXI-B",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "joined-aps/joined-ap/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/joined-aps/joined-ap/state/opstate": 1,
								"cisco/joined-aps/joined-ap/state/uptime":  1295277,
							},
						},
						{
							Name:      "joined-aps/joined-ap/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"eth-mac":           "00:00:00:00:00:01",
								"serial":            "ABCD1234",
								"model":             "C9105AXI-B",
							},
							Values: map[string]any{
								"cisco/joined-aps/joined-ap/state/info": 1,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/version": 17,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/release": 12,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/maint": 5,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/build": 41,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"version":           "17",
								"release":           "12",
								"maint":             "5",
								"build":             "41",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/info": 1,
							},
						},
					},
				},
				{
					name: "geolocation-oper-data/ap-geo-loc-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                 "wlc.example.org",
								"subscription-name":      "ap",
								"ap-geo-loc-data_ap-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/longitude": -121.93233,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                 "wlc.example.org",
								"subscription-name":      "ap",
								"ap-geo-loc-data_ap-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/latitude": 37.41199,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"latitude":          "37.41199",
								"longitude":         "-121.93233",
							},
							Values: map[string]any{
								"cisco/geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/info": 1,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/oper-data/ap-sys-stats/cpu-usage success",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 2,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 2,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/oper-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/external-module-data/xm-data/xm/numeric-id": 0,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/external-module-data/xm-data/xm/max-power": 0,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/version": 17,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/static-info/board-data/wtp-serial-num": "ABCD1234",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/static-info/board-data/ap-sys-info/mem-size": 1989632,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/country-code": "US ",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/static-info/board-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"serial-number":     "ABCD1234",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/device-detail/static-info/board-data/wtp-serial-num/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"version":           "17",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/device-detail/wtp-version/sw-ver/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/device-detail/static-info/board-data/ap-sys-info",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/device-detail/static-info/board-data/ap-sys-info/mem-size": 1989632,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data/external-module-data/xm-data/xm",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"numeric-id":        "0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/external-module-data/xm-data/xm/max-power": 0,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"country-code":      "US",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-data/country-code/info": 1,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/oper-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1763607104073103000,
							Tags: map[string]string{
								"source":              "wlc.example.org",
								"subscription-name":   "ap",
								"capwap-pkts_wtp-mac": "00:00:00:00:00:00",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-pkts/cntrl-pkts": 18,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/capwap-pkts",
							Timestamp: 1763607104073103000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-mac":            "00:00:00:00:00:00",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/capwap-pkts/cntrl-pkts": 18,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/ethernet-if-stats",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "0",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name": "GigabitEthernet0",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "0",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/rx-pkts": 14006461,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "0",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status": "oper-state-up",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "0",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name": "GigabitEthernet0",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "0",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/rx-pkts": 23928361,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "1",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name": "LAN1",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "1",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/rx-pkts": 0,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                     "wlc.example.org",
								"subscription-name":          "ap",
								"ethernet-if-stats_if-index": "1",
								"ethernet-if-stats_wtp-mac":  "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status": "oper-state-down",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"if-index":          "0",
								"if-name":           "GigabitEthernet0",
								"oper-status":       "oper-state-up",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/ethernet-if-stats/oper-status/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"if-index":          "0",
								"if-name":           "GigabitEthernet0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/ethernet-if-stats/rx-pkts":     14006461,
								"cisco/access-point-oper-data/ethernet-if-stats/oper-status": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
								"if-index":          "0",
								"if-name":           "GigabitEthernet0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/ethernet-if-stats/rx-pkts": 23928361,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
								"if-index":          "1",
								"if-name":           "LAN1",
								"oper-status":       "oper-state-down",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/ethernet-if-stats/oper-status/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
								"if-index":          "1",
								"if-name":           "LAN1",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/ethernet-if-stats/rx-pkts":     0,
								"cisco/access-point-oper-data/ethernet-if-stats/oper-status": 0,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/radio-oper-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "ap",
								"radio-oper-data_radio-slot-id": "0",
								"radio-oper-data_wtp-mac":       "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data/current-band-id": 0,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "ap",
								"radio-oper-data_radio-slot-id": "0",
								"radio-oper-data_wtp-mac":       "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data/current-active-band": "dot11-2-dot-4-ghz-band",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
								"band-id":           "0",
								"band":              "dot11-2-dot-4-ghz-band",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/radio-oper-data/current-active-band/info": 1,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data/curr-freq",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "ap",
								"radio-oper-data_radio-slot-id": "0",
								"radio-oper-data_wtp-mac":       "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data/curr-freq": 6,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
								"channel":           "6",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data/curr-freq/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/radio-oper-data/phy-ht-cfg/cfg-data/curr-freq": 6,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/radio-oper-stats",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"radio-oper-stats_slot-id": "0",
								"radio-oper-stats_ap-mac":  "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats/retry-count": 79895,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"radio-oper-stats_slot-id": "0",
								"radio-oper-stats_ap-mac":  "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats/ap-radio-stats/last-ts": "1970-01-01T00:00:00.000000Z",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/radio-oper-stats/retry-count": 79895,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-stats/ap-radio-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/radio-oper-stats/ap-radio-stats/last-ts": int64(
									0,
								),
							},
						},
					},
				},
				{
					name: "rrm-oper-data/rrm-measurement/noise/noise/noise-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap-json",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "ap-json",
								"rrm-measurement_radio-slot-id": "0",
								"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":   2,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise":  -95,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/chan":  11,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/noise": -94,
							},
						},
						{
							Name:      "ap-json",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "ap-json",
								"rrm-measurement_radio-slot-id": "1",
								"rrm-measurement_wtp-mac":       "01:00:00:00:00:02",
							},
							Values: map[string]any{
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":   134,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise":  -84,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/chan":  150,
								"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/noise": -85,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap-json",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
								"channel":           "2",
								"idx":               "1",
							},
							Values: map[string]any{
								"cisco/rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap-json",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "0",
								"channel":           "11",
								"idx":               "10",
							},
							Values: map[string]any{
								"cisco/rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap-json",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
								"slot-id":           "1",
								"channel":           "134",
								"idx":               "1",
							},
							Values: map[string]any{
								"cisco/rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -84,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap-json",
								"site-tag":          "Site2",
								"location-name":     "Location2",
								"ap-name":           "ap02",
								"ap-mac":            "01:00:00:00:00:02",
								"slot-id":           "1",
								"channel":           "150",
								"idx":               "10",
							},
							Values: map[string]any{
								"cisco/rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -85,
							},
						},
					},
				},
				{
					name: "client-oper-data/dot11-oper-data",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ap-mac-address": "00:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-wlan-id": 1,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-ap-slot-id": 1,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-wifi/wpa-version": "wpa2",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-wifi/cipher-suite": "ccmp-aes",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "1",
								"wlan-id":           "1",
								"ssid":              "ExampleSSID",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "1",
								"wlan-id":           "1",
								"ssid":              "ExampleSSID",
								"client-mac":        "02:00:00:00:00:01",
								"channel":           "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-wifi",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"site-tag":          "Site1",
								"location-name":     "Location1",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
								"slot-id":           "1",
								"wlan-id":           "1",
								"ssid":              "ExampleSSID",
								"client-mac":        "02:00:00:00:00:01",
								"wpa-version":       "wpa2",
								"cipher-suite":      "ccmp-aes",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/ms-wifi/info": 1,
							},
						},
					},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map delete",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Deletes: []string{
								"/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map delete",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Deletes: []string{
								"/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "client-oper-data/dc-info",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"dc-info_client-mac": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info/device-type": "Un-Classified Device",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"dc-info_client-mac": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info/confidence-level": 0,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"dc-info_client-mac": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info/classified-time": "2025-09-23T22:59:43.000000Z",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"device-type":       "Un-Classified Device",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dc-info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dc-info/confidence-level": 0,
								"cisco/client-oper-data/dc-info/classified-time": int64(
									1758668383,
								),
							},
						},
					},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map delete",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Deletes: []string{
								"/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "access-point-oper-data/ap-name-mac-map delete",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Deletes: []string{
								"/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name: "info metrics",
			calls: []testCall{
				{
					name: "single info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
								"/system/config/domain":   "example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
								"domain":            "example.org",
							},
							Values: map[string]any{
								"cisco/system/config/info": 1,
							},
						},
					},
				},
				{
					name: "individual info metrics",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
								"/system/state/hostname":  "wlc.example.org",
								"/system/state/boot-time": 1755731936,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
						{
							Name:      "system/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/state/info": 1,
							},
						},
						{
							Name:      "system/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/system/state/boot-time": 1755731936,
							},
						},
					},
				},
				{
					name: "device-inventory info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "wlc",
								"device-inventory_hw-type": "hw-type-chassis",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-type": "hw-type-chassis",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 0,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/part-number": "C9800-CL-K9",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-type":           "hw-type-chassis",
								"hw-dev-index":      "0",
								"version":           "V00",
								"part-number":       "C9800-CL-K9",
								"field-replaceable": "false",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "delete with values",
			calls: []testCall{
				{
					name: "delete with info metrics",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
								"/system/state/hostname":  "wlc.example.org",
							},
							Deletes: []string{
								"/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
						{
							Name:      "system/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/state/hostname/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "offline access point",
			calls: []testCall{
				{
					name: "ap-global-oper-data/ap-join-stats",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":                "wlc.example.org",
								"subscription-name":     "ap",
								"ap-join-stats_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info/ap-ethernet-mac": "00:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":                "wlc.example.org",
								"subscription-name":     "ap",
								"ap-join-stats_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info/ap-name": "ap01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":                "wlc.example.org",
								"subscription-name":     "ap",
								"ap-join-stats_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info/is-joined": false,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-mac":            "01:00:00:00:00:01",
								"ap-name":           "ap01",
								"eth-mac":           "00:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/ap-global-oper-data/ap-join-stats/ap-join-info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-ap-global-oper:ap-global-oper-data/ap-join-stats/ap-join-info",
							Timestamp: 1764291790855619000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-mac":            "01:00:00:00:00:01",
								"ap-name":           "ap01",
							},
							Values: map[string]any{
								"cisco/ap-global-oper-data/ap-join-stats/ap-join-info/is-joined": false,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})
	err := c.Init(map[string]any{})
	require.NoError(t, err)

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			for _, call := range test.calls {
				actual := c.Apply(call.input...)

				assert.ElementsMatch(t, call.expected, actual, call.name)
			}
		})
	}
}

func TestValueAllowBlockList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		allowlist []string
		blocklist []string
		calls     []testCall
	}

	testCases := []testCase{
		{
			name:      "empty allowlist/blocklist",
			allowlist: []string{},
			blocklist: []string{},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name:      "root allowlist",
			allowlist: []string{"/system"},
			blocklist: []string{},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name:      "exact match allowlist",
			allowlist: []string{"system/config/hostname"},
			blocklist: []string{},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name:      "root blocklist",
			allowlist: []string{},
			blocklist: []string{"/system"},
			calls: []testCall{
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:      "exact match blocklist",
			allowlist: []string{},
			blocklist: []string{"system/config/hostname"},
			calls: []testCall{
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:      "root blocklist",
			allowlist: []string{},
			blocklist: []string{"/system"},
			calls: []testCall{
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:      "multiple allowlist",
			allowlist: []string{"system/state", "/system/config/", "system/config/"},
			blocklist: []string{},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
				{
					name: "not allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"system/dns/servers/server/config/port": 53,
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/state/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/state/hostname/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name:      "multiple blocklist",
			allowlist: []string{},
			blocklist: []string{"system/state", "/system/config/", "system/config/"},
			calls: []testCall{
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "not blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"system/dns/servers/server/config/port": 53,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/dns/servers/server/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/system/dns/servers/server/config/port": 53,
							},
						},
					},
				},
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/state/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:      "allowlist and blocklist",
			allowlist: []string{"/system/"},
			blocklist: []string{"system/state"},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/config/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hostname":          "wlc.example.org",
							},
							Values: map[string]any{
								"cisco/system/config/hostname/info": 1,
							},
						},
					},
				},
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"system/dns/servers/server/config/port": 53,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "system/dns/servers/server/config",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/system/dns/servers/server/config/port": 53,
							},
						},
					},
				},
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"/system/state/hostname": "wlc.example.org",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name: "allowlist and blocklist with yang namespace",
			allowlist: []string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware",
			},
			blocklist: []string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
			},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/software-version": "Cisco IOS Software [Dublin], C9800-CL Software (C9800-CL-K9_IOSXE), Version 17.12.5, RELEASE SOFTWARE (fc5)\nTechnical Support: http://www.cisco.com/techsupport\nCopyright (c) 1986-2025 by Cisco Systems, Inc.\nCompiled Fri 14-Mar-25 02:50 by mcpre",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"software-version":  "Cisco IOS Software [Dublin], C9800-CL Software (C9800-CL-K9_IOSXE), Version 17.12.5, RELEASE SOFTWARE (fc5)\nTechnical Support: http://www.cisco.com/techsupport\nCopyright (c) 1986-2025 by Cisco Systems, Inc.\nCompiled Fri 14-Mar-25 02:50 by mcpre",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/software-version/info": 1,
							},
						},
					},
				},
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/part-number": "C9800-CL-K9",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:      "global allowlist",
			allowlist: []string{"/"},
			blocklist: []string{},
			calls: []testCall{
				{
					name: "allowed",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/part-number": "C9800-CL-K9",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"part-number":       "C9800-CL-K9",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name:      "global blocklist",
			allowlist: []string{},
			blocklist: []string{"/"},
			calls: []testCall{
				{
					name: "blocked",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/part-number": "C9800-CL-K9",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.ValueAllowlist = &test.allowlist
			c.ValueBlocklist = &test.blocklist
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for i, k := range test.allowlist {
					for _, x := range allYangPathCombinations(k, false) {
						allowlist := []string{}
						allowlist = append(allowlist, test.allowlist...)
						allowlist[i] = x

						c.ValueAllowlist = &allowlist
						c.ValueBlocklist = &test.blocklist
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}

				for i, k := range test.blocklist {
					for _, x := range allYangPathCombinations(k, false) {
						blocklist := []string{}
						blocklist = append(blocklist, test.blocklist...)
						blocklist[i] = x

						c.ValueAllowlist = &test.allowlist
						c.ValueBlocklist = &blocklist
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestTagRenames(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		tagrenames map[string]map[string]string
		calls      []testCall
	}

	testCases := []testCase{
		{
			name:       "empty",
			tagrenames: map[string]map[string]string{},
			calls: []testCall{
				{
					name: "no rename",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ms-mac-address":    "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ms-mac-address":    "02:00:00:00:00:01",
								"current-channel":   "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			tagrenames: map[string]map[string]string{
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data": {
					"ms-mac-address": "client-mac",
				},
			},
			calls: []testCall{
				{
					name: "rename ms-mac-address",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"current-channel":   "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "overlapping",
			tagrenames: map[string]map[string]string{
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data": {
					"ms-mac-address": "client-mac",
				},
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data": {
					"vap-ssid": "ssid",
				},
			},
			calls: []testCall{
				{
					name: "rename ms-mac-address and vap-ssid",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"current-channel":   "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "global match",
			tagrenames: map[string]map[string]string{
				"": {
					"ms-mac-address": "client-mac",
					"vap-ssid":       "ssid",
				},
			},
			calls: []testCall{
				{
					name: "rename ms-mac-address and vap-ssid",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"current-channel":   "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.tagRenames = test.tagrenames
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for k, v := range test.tagrenames {
					for _, x := range allYangPathCombinations(k, true) {
						tagrenames := map[string]map[string]string{}
						maps.Copy(tagrenames, test.tagrenames)
						delete(tagrenames, k)
						if _, exists := tagrenames[x]; exists {
							maps.Copy(tagrenames[x], v)
						} else {
							tagrenames[x] = v
						}

						c.tagRenames = tagrenames
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestValueTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		valuetags map[string][]string
		calls     []testCall
	}

	testCases := []testCase{
		{
			name:      "empty",
			valuetags: map[string][]string{},
			calls: []testCall{
				{
					name: "no value tags",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"channel":           "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			valuetags: map[string][]string{
				"/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/": {
					"vap-ssid",
				},
			},
			calls: []testCall{
				{
					name: "ssid tag",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"channel":           "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "overlapping",
			valuetags: map[string][]string{
				"/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/": {
					"vap-ssid",
				},
				"/Cisco-IOS-XE-wireless-client-oper:client-oper-data/": {
					"ms-ap-slot-id",
				},
			},
			calls: []testCall{
				{
					name: "ssid and slot-id tags",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-ap-slot-id": 1,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"slot-id":           "1",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"slot-id":           "1",
								"channel":           "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "global match",
			valuetags: map[string][]string{
				"/": {
					"vap-ssid",
					"ms-ap-slot-id",
				},
			},
			calls: []testCall{
				{
					name: "ssid and slot-id tags",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/vap-ssid": "ExampleSSID",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                         "wlc.example.org",
								"subscription-name":              "ap",
								"dot11-oper-data_ms-mac-address": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data/ms-ap-slot-id": 1,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"slot-id":           "1",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel": 140,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dot11-oper-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"ssid":              "ExampleSSID",
								"slot-id":           "1",
								"channel":           "140",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dot11-oper-data/current-channel/info": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.valueTags = test.valuetags
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for k, v := range test.valuetags {
					for _, x := range allYangPathCombinations(k, true) {
						valuetags := map[string][]string{}
						maps.Copy(valuetags, test.valuetags)
						delete(valuetags, k)
						if curr, exists := valuetags[x]; exists {
							valuetags[x] = append(curr, v...)
						} else {
							valuetags[x] = v
						}

						c.valueTags = valuetags
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestEnumMappings(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		enummappings map[string]map[string]map[string]int
		calls        []testCall
	}

	testCases := []testCase{
		{
			name:         "empty",
			enummappings: map[string]map[string]map[string]int{},
			calls: []testCall{
				{
					name: "no enum mapping",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status-desc": "status-desc-ok",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
								"status-desc":       "status-desc-ok",
							},
							Values: map[string]any{
								"cisco/components/component/state/status-desc/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			enummappings: map[string]map[string]map[string]int{
				"Cisco-IOS-XE-platform-oper:components/component/state": {
					"status-desc": {
						"status-desc-ok": 1,
					},
				},
			},
			calls: []testCall{
				{
					name: "status-desc enum",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status-desc": "status-desc-ok",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/status-desc": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "overlapping",
			enummappings: map[string]map[string]map[string]int{
				"Cisco-IOS-XE-platform-oper:components/component/state": {
					"status-desc": {
						"status-desc-ok": 1,
					},
				},
				"Cisco-IOS-XE-platform-oper:components/component": {
					"status": {
						"status-active": 1,
					},
				},
			},
			calls: []testCall{
				{
					name: "status and status-desc enums",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status-desc": "status-desc-ok",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status": "status-active",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/status":      1,
								"cisco/components/component/state/status-desc": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "global match",
			enummappings: map[string]map[string]map[string]int{
				"": {
					"status-desc": {
						"status-desc-ok": 1,
					},
					"status": {
						"status-active": 1,
					},
				},
			},
			calls: []testCall{
				{
					name: "status and status-desc enums",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status-desc": "status-desc-ok",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status": "status-active",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/status":      1,
								"cisco/components/component/state/status-desc": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.enumMappings = test.enummappings
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for k, v := range test.enummappings {
					for _, x := range allYangPathCombinations(k, true) {
						enummappings := map[string]map[string]map[string]int{}
						maps.Copy(enummappings, test.enummappings)
						delete(enummappings, k)
						if _, exists := enummappings[x]; exists {
							maps.Copy(enummappings[x], v)
						} else {
							enummappings[x] = v
						}

						c.enumMappings = enummappings
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestIntegerInfoMetrics(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name               string
		integerinfometrics map[string][]string
		calls              []testCall
	}

	testCases := []testCase{
		{
			name:               "empty",
			integerinfometrics: map[string][]string{},
			calls: []testCall{
				{
					name: "no info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"policy-data_mac":   "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			integerinfometrics: map[string][]string{
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data": {
					"res-vlan-id",
				},
			},
			calls: []testCall{
				{
					name: "res-vlan-id info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"policy-data_mac":   "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"res-vlan-id":       "10",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "overlapping",
			integerinfometrics: map[string][]string{
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data": {
					"res-vlan-id",
				},
				"Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info": {
					"classified-time",
				},
			},
			calls: []testCall{
				{
					name: "res-vlan-id and classified-time info metrics",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"policy-data_mac":   "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"dc-info_client-mac": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info/classified-time": "2025-09-23T22:59:43.000000Z",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"res-vlan-id":       "10",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"classified-time":   "1758668383",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dc-info/classified-time/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "global match",
			integerinfometrics: map[string][]string{
				"": {
					"res-vlan-id",
					"classified-time",
				},
			},
			calls: []testCall{
				{
					name: "res-vlan-id and classified-time info metrics",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"policy-data_mac":   "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":             "wlc.example.org",
								"subscription-name":  "ap",
								"dc-info_client-mac": "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info/classified-time": "2025-09-23T22:59:43.000000Z",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id": 10,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/policy-data",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"res-vlan-id":       "10",
							},
							Values: map[string]any{
								"cisco/client-oper-data/policy-data/res-vlan-id/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-wireless-client-oper:client-oper-data/dc-info",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"client-mac":        "02:00:00:00:00:01",
								"classified-time":   "1758668383",
							},
							Values: map[string]any{
								"cisco/client-oper-data/dc-info/classified-time/info": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.integerInfoMetrics = test.integerinfometrics
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for k, v := range test.integerinfometrics {
					for _, x := range allYangPathCombinations(k, true) {
						integerinfometrics := map[string][]string{}
						maps.Copy(integerinfometrics, test.integerinfometrics)
						delete(integerinfometrics, k)
						if curr, exists := integerinfometrics[x]; exists {
							integerinfometrics[x] = append(curr, v...)
						} else {
							integerinfometrics[x] = v
						}

						c.integerInfoMetrics = integerinfometrics
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestValueAndInfoMetrics(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                string
		valueandinfometrics map[string][]string
		calls               []testCall
	}

	testCases := []testCase{
		{
			name:                "empty",
			valueandinfometrics: map[string][]string{},
			calls: []testCall{
				{
					name: "only info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"reason-severity":   "normal",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/reason-severity/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			valueandinfometrics: map[string][]string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data": {
					"reason-severity",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"reason-severity":   "normal",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/reason-severity/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
							},
						},
					},
				},
			},
		},
		{
			name: "overlapping",
			valueandinfometrics: map[string][]string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data": {
					"reason-severity",
				},
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware": {
					"rommon-version",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value for reason-severity and rommon-version",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/rommon-version": "IOS-XE ROMMON",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"reason-severity":   "normal",
								"rommon-version":    "IOS-XE ROMMON",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
								"cisco/device-hardware-data/device-hardware/device-system-data/rommon-version":  "IOS-XE ROMMON",
							},
						},
					},
				},
			},
		},
		{
			name: "global match",
			valueandinfometrics: map[string][]string{
				"": {
					"reason-severity",
					"rommon-version",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value for reason-severity and rommon-version",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data/rommon-version": "IOS-XE ROMMON",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"reason-severity":   "normal",
								"rommon-version":    "IOS-XE ROMMON",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-system-data",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-system-data/reason-severity": "normal",
								"cisco/device-hardware-data/device-hardware/device-system-data/rommon-version":  "IOS-XE ROMMON",
							},
						},
					},
				},
			},
		},
		{
			name: "timestamp",
			valueandinfometrics: map[string][]string{
				"Cisco-IOS-XE-platform-oper:components/component/state": {
					"mfg-date",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/mfg-date": "2000-01-01T00:00:00+00:00",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
								"mfg-date":          "2000-01-01T00:00:00+00:00",
							},
							Values: map[string]any{
								"cisco/components/component/state/mfg-date/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/mfg-date": int64(
									946684800,
								),
							},
						},
					},
				},
			},
		},
		{
			name: "enum",
			valueandinfometrics: map[string][]string{
				"Cisco-IOS-XE-platform-oper:components/component/state": {
					"status",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status": "status-active",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
								"status":            "status-active",
							},
							Values: map[string]any{
								"cisco/components/component/state/status/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/status": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "enum and timestamp",
			valueandinfometrics: map[string][]string{
				"Cisco-IOS-XE-platform-oper:components/component/state": {
					"status",
					"mfg-date",
				},
			},
			calls: []testCall{
				{
					name: "info metric and value",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/mfg-date": "2000-01-01T00:00:00+00:00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"component_cname":   "SlotF0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-platform-oper:components/component/state/status": "status-active",
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
								"mfg-date":          "2000-01-01T00:00:00+00:00",
							},
							Values: map[string]any{
								"cisco/components/component/state/mfg-date/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
								"status":            "status-active",
							},
							Values: map[string]any{
								"cisco/components/component/state/status/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-platform-oper:components/component/state",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"cname":             "SlotF0",
							},
							Values: map[string]any{
								"cisco/components/component/state/mfg-date": int64(
									946684800,
								),
								"cisco/components/component/state/status": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.enumMappings = map[string]map[string]map[string]int{
				"Cisco-IOS-XE-platform-oper:components": {
					"status": {
						"status-active": 1,
					},
				},
			}
			c.valueAndInfoMetrics = test.valueandinfometrics
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for k, v := range test.valueandinfometrics {
					for _, x := range allYangPathCombinations(k, true) {
						valueandinfometrics := map[string][]string{}
						maps.Copy(valueandinfometrics, test.valueandinfometrics)
						delete(valueandinfometrics, k)
						if curr, exists := valueandinfometrics[x]; exists {
							valueandinfometrics[x] = append(curr, v...)
						} else {
							valueandinfometrics[x] = v
						}

						c.valueAndInfoMetrics = valueandinfometrics
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestMixedInfoMetrics(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name             string
		mixedinfometrics []string
		calls            []testCall
	}

	testCases := []testCase{
		{
			name:             "empty",
			mixedinfometrics: []string{},
			calls: []testCall{
				{
					name: "individual metrics",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 0,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-dev-index":      "0",
								"hw-type":           "hw-type-chassis",
								"version":           "V00",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-dev-index":      "0",
								"hw-type":           "hw-type-chassis",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/hw-dev-index":      0,
								"cisco/device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
					},
				},
			},
		},
		{
			name: "single",
			mixedinfometrics: []string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
			},
			calls: []testCall{
				{
					name: "single info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 0,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-dev-index":      "0",
								"hw-type":           "hw-type-chassis",
								"version":           "V00",
								"field-replaceable": "false",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
					},
				},
			},
		},
		{
			name: "multiple",
			mixedinfometrics: []string{
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
				"Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center",
			},
			calls: []testCall{
				{
					name: "single info metric",
					input: []*formatters.EventMsg{
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/hw-dev-index": 0,
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/version": "V00",
							},
						},
						{
							Name:      "wlc",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":                        "wlc.example.org",
								"subscription-name":             "wlc",
								"device-inventory_hw-type":      "hw-type-chassis",
								"device-inventory_hw-dev-index": "0",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory/field-replaceable": false,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
							Timestamp: 1757461050645191000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "wlc",
								"hw-dev-index":      "0",
								"hw-type":           "hw-type-chassis",
								"version":           "V00",
								"field-replaceable": "false",
							},
							Values: map[string]any{
								"cisco/device-hardware-data/device-hardware/device-inventory/info": 1,
							},
						},
					},
				},
				{
					name: "single info metric",
					input: []*formatters.EventMsg{
						{

							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/longitude": -121.93233,
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/latitude": 37.41199,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-geolocation-oper:geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"latitude":          "37.41199",
								"longitude":         "-121.93233",
							},
							Values: map[string]any{
								"cisco/geolocation-oper-data/ap-geo-loc-data/loc/ellipse/center/info": 1,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.MixedInfoMetrics = &test.mixedinfometrics
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for i, k := range test.mixedinfometrics {
					for _, x := range allYangPathCombinations(k, false) {
						mixedinfometrics := []string{}
						mixedinfometrics = append(mixedinfometrics, test.mixedinfometrics...)
						mixedinfometrics[i] = x

						c.MixedInfoMetrics = &mixedinfometrics
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}

func TestApTags(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		aptags       []string
		bypassaptags []string
		calls        []testCall
	}

	testCases := []testCase{
		{
			name:         "empty aptags",
			aptags:       []string{},
			bypassaptags: []string{},
			calls: []testCall{
				{
					name: "metric rejected for missing ap-name",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
			},
		},
		{
			name:         "minimal aptags",
			aptags:       []string{"ap-mac", "ap-name"},
			bypassaptags: []string{},
			calls: []testCall{
				{
					name: "access-point-oper-data/ap-name-mac-map",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/wtp-mac": "01:00:00:00:00:01",
							},
						},
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":                   "wlc.example.org",
								"subscription-name":        "ap",
								"ap-name-mac-map_wtp-name": "ap01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map/eth-mac": "00:00:00:00:00:01",
							},
						},
					},
					expected: []*formatters.EventMsg{},
				},
				{
					name: "metric accepted",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"ap-name":           "ap01",
								"ap-mac":            "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
				},
			},
		},
		{
			name:   "bypassed aptags",
			aptags: []string{"ap-mac", "ap-name"},
			bypassaptags: []string{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data",
			},
			calls: []testCall{
				{
					name: "metric accepted",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"wtp-mac":           "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
				},
			},
		},
		{
			name:         "globally bypassed aptags",
			aptags:       []string{"ap-mac", "ap-name"},
			bypassaptags: []string{"/"},
			calls: []testCall{
				{
					name: "metric accepted",
					input: []*formatters.EventMsg{
						{
							Name:      "ap",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"oper-data_wtp-mac": "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
					expected: []*formatters.EventMsg{
						{
							Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/oper-data/ap-sys-stats",
							Timestamp: 1757474904984599000,
							Tags: map[string]string{
								"source":            "wlc.example.org",
								"subscription-name": "ap",
								"wtp-mac":           "01:00:00:00:00:01",
							},
							Values: map[string]any{
								"cisco/access-point-oper-data/oper-data/ap-sys-stats/cpu-usage": 6,
							},
						},
					},
				},
			},
		},
	}

	c := prometheusProcessor{
		Debug:               true,
		MetricPrefix:        "cisco",
		CiscoWLCs:           &[]string{"wlc.example.org"},
		ApTags:              apTags,
		BypassApTags:        bypassApTags,
		tagRenames:          tagRenames,
		valueTags:           valueTags,
		enumMappings:        enumMappings,
		integerInfoMetrics:  integerInfoMetrics,
		valueAndInfoMetrics: valueAndInfoMetrics,
		MixedInfoMetrics:    mixedInfoMetrics,
		ValueBlocklist:      valueBlocklist,
	}
	c.WithTargets(map[string]*types.TargetConfig{
		"wlc.example.org": {
			Name:    "wlc.example.org",
			Address: "wlc.example.org:50052",
		},
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			c.ApTags = &test.aptags
			c.BypassApTags = &test.bypassaptags
			err := c.Init(map[string]any{})
			require.NoError(t, err)

			for _, call := range test.calls {
				actual := c.Apply(call.input...)
				assert.ElementsMatch(t, call.expected, actual, call.name)

				for i, k := range test.bypassaptags {
					for _, x := range allYangPathCombinations(k, false) {
						bypassaptags := []string{}
						bypassaptags = append(bypassaptags, test.bypassaptags...)
						bypassaptags[i] = x

						c.BypassApTags = &bypassaptags
						err := c.Init(map[string]any{})
						require.NoError(t, err)

						actual := c.Apply(call.input...)
						assert.ElementsMatch(t, call.expected, actual, call.name)
					}
				}
			}
		})
	}
}
