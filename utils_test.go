package main

import (
	"fmt"
	"maps"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/itchyny/gojq"
	"github.com/openconfig/gnmic/pkg/formatters"
	"github.com/stretchr/testify/assert"
)

func TestStripYangOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"full path with origin and namespace",
			"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name",
			"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name",
		},
		{
			"root path with origin",
			"rfc7951:/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"root path with namespace and leading slash",
			"/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"root path with namespace and leading/trailing slash",
			"/Cisco-IOS-XE-process-cpu-oper:cpu-usage/",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"root path with namespace",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"openconfig full path with origin and namespace",
			"openconfig:/openconfig:system/config/hostname",
			"openconfig:system/config/hostname",
		},
		{
			"openconfig full path with origin, namespace and leading slash",
			"/openconfig:system/config/hostname",
			"openconfig:system/config/hostname",
		},
		{
			"openconfig full path with origin, namespace and leading/trailing slash",
			"/openconfig:system/config/hostname/",
			"openconfig:system/config/hostname",
		},
		{
			"openconfig full path with namespace",
			"openconfig:system/config/hostname",
			"openconfig:system/config/hostname",
		},
		{
			"openconfig full path with origin",
			"openconfig:/system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with leading slash",
			"/system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with leading/trailing slash",
			"/system/config/hostname/",
			"system/config/hostname",
		},
		{
			"openconfig full path",
			"system/config/hostname",
			"system/config/hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := stripYangPathOrigin(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStripYangPathNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"root path with origin and namespace",
			"rfc7951:/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"cpu-usage",
		},
		{
			"root path with namespace and slash",
			"/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"cpu-usage",
		},
		{
			"root path with namespace",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"cpu-usage",
		},
		{
			"openconfig full path with origin and namespace",
			"openconfig:/openconfig:system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with namespace and slash",
			"/openconfig:system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with namespace",
			"openconfig:system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with origin",
			"openconfig:/system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path with slash",
			"/system/config/hostname",
			"system/config/hostname",
		},
		{
			"openconfig full path",
			"system/config/hostname",
			"system/config/hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := stripYangPathNamespace(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"root path with origin and namespace",
			"rfc7951:/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"root path with namespace and slash",
			"/Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"root path with namespace",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
			"Cisco-IOS-XE-process-cpu-oper:cpu-usage",
		},
		{
			"openconfig full path with origin and namespace",
			"openconfig:/openconfig:system/config/hostname",
			"openconfig:system",
		},
		{
			"openconfig full path with namespace and slash",
			"/openconfig:system/config/hostname",
			"openconfig:system",
		},
		{
			"openconfig full path with namespace",
			"openconfig:system/config/hostname",
			"openconfig:system",
		},
		{
			"openconfig full path with origin",
			"openconfig:/system/config/hostname",
			"system",
		},
		{
			"openconfig full path with slash",
			"/system/config/hostname",
			"system",
		},
		{
			"openconfig full path",
			"system/config/hostname",
			"system",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := getYangPathRoot(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStripTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name: "single",
			input: map[string]string{
				"foo_bar": "baz",
			},
			expected: map[string]string{
				"bar": "baz",
			},
		},
		{
			name: "multiple",
			input: map[string]string{
				"foo_bar": "baz",
				"a_b":     "c",
				"d":       "e",
			},
			expected: map[string]string{
				"bar": "baz",
				"b":   "c",
				"d":   "e",
			},
		},
		{
			name: "complex",
			input: map[string]string{
				"ethernet-if-stats_if-index": "1",
				"ethernet-if-stats_wtp-mac":  "aa:bb:cc:dd:ee:ff",
				"source":                     "example.org",
				"subscription-name":          "example",
			},
			expected: map[string]string{
				"if-index":          "1",
				"wtp-mac":           "aa:bb:cc:dd:ee:ff",
				"source":            "example.org",
				"subscription-name": "example",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := stripTags(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestAddTag(t *testing.T) {
	t.Parallel()

	type input struct {
		tags  map[string]string
		tag   string
		value string
	}

	type expected struct {
		tags map[string]string
		err  error
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			"empty tags",
			input{
				tags:  map[string]string{},
				tag:   "test",
				value: "foo",
			},
			expected{
				tags: map[string]string{
					"test": "foo",
				},
				err: nil,
			},
		},
		{
			"exisiting tags",
			input{
				tags: map[string]string{
					"foo": "bar",
				},
				tag:   "test",
				value: "foo",
			},
			expected{
				tags: map[string]string{
					"foo":  "bar",
					"test": "foo",
				},
				err: nil,
			},
		},
		{
			"overlapping tag with same value",
			input{
				tags: map[string]string{
					"foo": "bar",
				},
				tag:   "foo",
				value: "bar",
			},
			expected{
				tags: map[string]string{
					"foo": "bar",
				},
				err: nil,
			},
		},
		{
			"overlapping tag with different value",
			input{
				tags: map[string]string{
					"foo": "bar",
				},
				tag:   "foo",
				value: "baz",
			},
			expected{
				tags: map[string]string{
					"foo": "bar",
				},
				err: fmt.Errorf(
					"tag foo exists with a different value (bar != baz): %w",
					ErrTagExists,
				),
			},
		},
		{
			"value with whitespace",
			input{
				tags:  map[string]string{},
				tag:   "foo",
				value: " baz ",
			},
			expected{
				tags: map[string]string{
					"foo": "baz",
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := addTag(test.input.tags, test.input.tag, test.input.value)

			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.tags, test.input.tags)
		})
	}
}

func TestStripValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "single",
			input: map[string]any{
				"foo": "bar",
			},
			expected: map[string]any{
				"cisco/foo": "bar",
			},
		},
		{
			name: "multiple",
			input: map[string]any{
				"foo": "bar",
				"a":   100,
			},
			expected: map[string]any{
				"cisco/foo": "bar",
				"cisco/a":   100,
			},
		},
		{
			name: "access-point-if-stats",
			input: map[string]any{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/duplex":       1,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name":      "LAN1",
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-crc":    0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-drops":  0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-errors": 0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-frames": 0,
			},
			expected: map[string]any{
				"cisco/access-point-oper-data/ethernet-if-stats/duplex":       1,
				"cisco/access-point-oper-data/ethernet-if-stats/if-name":      "LAN1",
				"cisco/access-point-oper-data/ethernet-if-stats/input-crc":    0,
				"cisco/access-point-oper-data/ethernet-if-stats/input-drops":  0,
				"cisco/access-point-oper-data/ethernet-if-stats/input-errors": 0,
				"cisco/access-point-oper-data/ethernet-if-stats/input-frames": 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := stripValues(test.input, "cisco")

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestMatchEvents(t *testing.T) {
	t.Parallel()

	type expected struct {
		matched   []*formatters.EventMsg
		unmatched []*formatters.EventMsg
	}

	tests := []struct {
		name      string
		condition string
		input     []*formatters.EventMsg
		expected  expected
	}{
		{
			"no condition",
			``,
			[]*formatters.EventMsg{
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
			expected{
				[]*formatters.EventMsg{
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
				[]*formatters.EventMsg{},
			},
		},
		{
			"include by hostname",
			`.tags.source == "wlc.example.org"`,
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc2.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"/system/config/hostname": "wlc2.example.org",
					},
				},
			},
			expected{
				[]*formatters.EventMsg{
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
				[]*formatters.EventMsg{
					{
						Name:      "wlc",
						Timestamp: 1757461050645191000,
						Tags: map[string]string{
							"source":            "wlc2.example.org",
							"subscription-name": "wlc",
						},
						Values: map[string]any{
							"/system/config/hostname": "wlc2.example.org",
						},
					},
				},
			},
		},
		{
			"exclude by value",
			`.values.["/system/config/hostname"] != "wlc.example.org"`,
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc2.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"/system/config/hostname": "wlc2.example.org",
					},
				},
			},
			expected{
				[]*formatters.EventMsg{
					{
						Name:      "wlc",
						Timestamp: 1757461050645191000,
						Tags: map[string]string{
							"source":            "wlc2.example.org",
							"subscription-name": "wlc",
						},
						Values: map[string]any{
							"/system/config/hostname": "wlc2.example.org",
						},
					},
				},
				[]*formatters.EventMsg{
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
			},
		},
		{
			"exclude all",
			`false`,
			[]*formatters.EventMsg{
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
			expected{
				[]*formatters.EventMsg{},
				[]*formatters.EventMsg{
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
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var condition *gojq.Code

			if test.condition != "" {
				q, err := gojq.Parse(test.condition)
				require.NoError(t, err)
				condition, err = gojq.Compile(q)
				require.NoError(t, err)
			}

			actualMatched, actualUnmatched, err := matchEvents(test.input, condition)
			require.NoError(t, err)

			assert.ElementsMatch(t, test.expected.matched, actualMatched)
			assert.ElementsMatch(t, test.expected.unmatched, actualUnmatched)
		})
	}
}

func TestSplitEventsByJsonArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []*formatters.EventMsg
		expected []*formatters.EventMsg
	}{
		{
			"single event",
			[]*formatters.EventMsg{
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
			[]*formatters.EventMsg{
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
		},
		{
			"multiple events",
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc.example.org",
					},
				},
			},
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc.example.org",
					},
				},
			},
		},
		{
			"single json array with one item",
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/noise": -95,
					},
				},
			},
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
			},
		},
		{
			"single json array with multiple items",
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/noise": -95,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise": -94,
					},
				},
			},
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "1",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
					},
				},
			},
		},
		{
			"multiple json array with multiple items",
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/noise": -95,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise": -94,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/foo.0/bar":          20,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/foo.1/bar":          30,
					},
				},
			},
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "1",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/foo/bar": 20,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "1",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/foo/bar": 30,
					},
				},
			},
		},
		{
			"multiple json array with multiple items",
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/noise": -95,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise": -94,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/chan":  153,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/noise": -95,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.11/chan":  154,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.11/noise": -94,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:02",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/chan":  140,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.10/noise": -85,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.11/chan":  144,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.11/noise": -84,
					},
				},
			},
			[]*formatters.EventMsg{
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "1",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "10",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  153,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "11",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  154,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:02",
						"json_idx":                      "10",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  140,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -85,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "1",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:02",
						"json_idx":                      "11",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  144,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -84,
					},
				},
			},
		},
		{
			"mixed flat and json array events",
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc.example.org",
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.0/noise": -95,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data.1/noise": -94,
					},
				},
			},
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc.example.org",
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "0",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  1,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -95,
					},
				},
				{
					Name:      "ap-json",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":                        "wlc.example.org",
						"subscription-name":             "ap-json",
						"rrm-measurement_radio-slot-id": "0",
						"rrm-measurement_wtp-mac":       "01:00:00:00:00:01",
						"json_idx":                      "1",
					},
					Values: map[string]any{
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/chan":  2,
						"/rfc7951:/Cisco-IOS-XE-wireless-rrm-oper:rrm-oper-data/rrm-measurement/noise/noise/noise-data/noise": -94,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := splitEventsByJSONArray(test.input)

			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}

func TestGroupEventsByTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []*formatters.EventMsg
		expected map[string]*formatters.EventMsg
	}{
		{
			"single event",
			[]*formatters.EventMsg{
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
			map[string]*formatters.EventMsg{
				"map[source:wlc.example.org subscription-name:wlc]": {
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
		},
		{
			"multiple events",
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc.example.org",
					},
				},
			},
			map[string]*formatters.EventMsg{
				"map[source:wlc.example.org subscription-name:wlc]": {
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
				},
			},
		},
		{
			"different tags",
			[]*formatters.EventMsg{
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
				{
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"subscription-name": "wlc",
						"source":            "wlc2.example.org",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc2.example.org",
					},
				},
			},
			map[string]*formatters.EventMsg{
				"map[source:wlc.example.org subscription-name:wlc]": {
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
				"map[source:wlc2.example.org subscription-name:wlc]": {
					Name:      "wlc",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc2.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"/system/state/hostname": "wlc2.example.org",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := groupEventsByTag(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEventToTree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *formatters.EventMsg
		expected map[string]*formatters.EventMsg
	}{
		{
			"single value",
			&formatters.EventMsg{
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
			map[string]*formatters.EventMsg{
				"system/config": {
					Name:      "system/config",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/config/hostname": "wlc.example.org",
					},
				},
			},
		},
		{
			"multiple values",
			&formatters.EventMsg{
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
			},
			map[string]*formatters.EventMsg{
				"system/config": {
					Name:      "system/config",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/config/hostname": "wlc.example.org",
					},
				},
				"system/state": {
					Name:      "system/state",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/state/hostname": "wlc.example.org",
					},
				},
			},
		},
		{
			"nested values",
			&formatters.EventMsg{
				Name:      "wlc",
				Timestamp: 1757461050645191000,
				Tags: map[string]string{
					"source":            "wlc.example.org",
					"subscription-name": "wlc",
				},
				Values: map[string]any{
					"/system/state/boot-time": 1755731936,
					"/system/config/hostname": "wlc.example.org",
					"/system/state/hostname":  "wlc.example.org",
					"/system/config/domain":   "example.org",
				},
			},
			map[string]*formatters.EventMsg{
				"system/config": {
					Name:      "system/config",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/config/hostname": "wlc.example.org",
						"system/config/domain":   "example.org",
					},
				},
				"system/state": {
					Name:      "system/state",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/state/hostname":  "wlc.example.org",
						"system/state/boot-time": 1755731936,
					},
				},
			},
		},
		{
			"multiple roots with nested values",
			&formatters.EventMsg{
				Name:      "wlc",
				Timestamp: 1757461050645191000,
				Tags: map[string]string{
					"source":            "wlc.example.org",
					"subscription-name": "wlc",
				},
				Values: map[string]any{
					"/foo/test":   1234,
					"/foo/child1": "test",
					"/bar/test":   5678,
					"/bar/child2": "test",
				},
			},
			map[string]*formatters.EventMsg{
				"foo": {
					Name:      "foo",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"foo/test":   1234,
						"foo/child1": "test",
					},
				},
				"bar": {
					Name:      "bar",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"bar/test":   5678,
						"bar/child2": "test",
					},
				},
			},
		},
		{
			"deeply nested values",
			&formatters.EventMsg{
				Name:      "wlc",
				Timestamp: 1757461050645191000,
				Tags: map[string]string{
					"source":            "wlc.example.org",
					"subscription-name": "wlc",
				},
				Values: map[string]any{
					"/foo/bar/baz/hi": 1234,
					"/foo/bar/test2":  3456,
					"/foo/bar/test1":  7890,
					"/foo/child1":     "test",
				},
			},
			map[string]*formatters.EventMsg{
				"foo": {
					Name:      "foo",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"foo/child1": "test",
					},
				},
				"foo/bar": {
					Name:      "foo/bar",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"foo/bar/test2": 3456,
						"foo/bar/test1": 7890,
					},
				},
				"foo/bar/baz": {
					Name:      "foo/bar/baz",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"foo/bar/baz/hi": 1234,
					},
				},
			},
		},
		{
			"delete",
			&formatters.EventMsg{
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
			map[string]*formatters.EventMsg{
				"system/config": {
					Name:      "system/config",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/config/hostname": "wlc.example.org",
					},
				},
				"system/state": {
					Name:      "system/state",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Values: map[string]any{
						"system/state/hostname": "wlc.example.org",
					},
				},
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map": {
					Name:      "Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
					Timestamp: 1757461050645191000,
					Tags: map[string]string{
						"source":            "wlc.example.org",
						"subscription-name": "wlc",
					},
					Deletes: []string{
						"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ap-name-mac-map",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := eventToTree(test.input)

			actual := maps.Collect(result.PrefixSearchIter(""))

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestRenameTags(t *testing.T) {
	t.Parallel()

	type input struct {
		yangPath   string
		tags       map[string]string
		tagRenames map[string]map[string]string
	}

	tests := []struct {
		name     string
		input    input
		expected map[string]string
	}{
		{
			"single",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data",
				map[string]string{
					"current-band-id": "0",
				},
				map[string]map[string]string{
					"access-point-oper-data": {
						"current-band-id": "band-id",
					},
				},
			},
			map[string]string{
				"band-id": "0",
			},
		},
		{
			"multiple",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/radio-oper-data",
				map[string]string{
					"wtp-mac":         "01:00:00:00:00:01",
					"current-band-id": "0",
					"radio-slot-id":   "1",
				},
				map[string]map[string]string{
					"access-point-oper-data/radio-oper-data": {
						"current-band-id": "band-id",
					},
					"access-point-oper-data": {
						"radio-slot-id": "slot-id",
					},
				},
			},
			map[string]string{
				"wtp-mac": "01:00:00:00:00:01",
				"band-id": "0",
				"slot-id": "1",
			},
		},
		{
			"none",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
				map[string]string{
					"if-index": "0",
				},
				map[string]map[string]string{
					"access-point-oper-data": {
						"current-band-id": "band-id",
					},
				},
			},
			map[string]string{
				"if-index": "0",
			},
		},
		{
			"most specific to least specific matching",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
				map[string]string{
					"if-index": "0",
				},
				map[string]map[string]string{
					"access-point-oper-data/ethernet-if-stats": {
						"if-index": "interface-id",
					},
					"access-point-oper-data": {
						"if-index": "if-id",
					},
				},
			},
			map[string]string{
				"interface-id": "0",
			},
		},
		{
			"less specific",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats",
				map[string]string{
					"if-index": "0",
				},
				map[string]map[string]string{
					"access-point-oper-data/ethernet-if-stats": {
						"if-id": "interface-id",
					},
					"access-point-oper-data": {
						"if-index": "if-id",
					},
				},
			},
			map[string]string{
				"if-id": "0",
			},
		},
		{
			"openconfig with leading slash",
			input{
				"/system/state",
				map[string]string{
					"system-id": "1",
				},
				map[string]map[string]string{
					"/system": {
						"system-id": "id",
					},
				},
			},
			map[string]string{
				"system-id": "1",
			},
		},
		{
			"openconfig with trailing slash",
			input{
				"/system/state",
				map[string]string{
					"system-id": "1",
				},
				map[string]map[string]string{
					"system/": {
						"system-id": "id",
					},
				},
			},
			map[string]string{
				"system-id": "1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := renameTags(
				test.input.yangPath,
				test.input.tags,
				test.input.tagRenames,
			)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestMergeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]map[string]any
	}{
		{
			name: "single",
			input: map[string]any{
				"foo": "bar",
			},
			expected: map[string]map[string]any{
				".": {"foo": "bar"},
			},
		},
		{
			name: "multiple",
			input: map[string]any{
				"foo/test1": "bar",
				"foo/test2": "baz",
				"a":         100,
			},
			expected: map[string]map[string]any{
				"foo": {"test1": "bar", "test2": "baz"},
				".":   {"a": 100},
			},
		},
		{
			name: "access-point-if-stats",
			input: map[string]any{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/duplex":       1,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name":      "LAN1",
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-crc":    0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-drops":  0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-errors": 0,
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-frames": 0,
			},
			expected: map[string]map[string]any{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats": {
					"duplex":       1,
					"if-name":      "LAN1",
					"input-crc":    0,
					"input-drops":  0,
					"input-errors": 0,
					"input-frames": 0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := mergeValues(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestRemoveValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		valuesToRemove []string
		valuesToKeep   []string
		values         map[string]any
		expected       map[string]any
	}{
		{
			name:           "remove none",
			valuesToRemove: []string{},
			valuesToKeep:   []string{},
			values: map[string]any{
				"foo": "bar",
			},
			expected: map[string]any{
				"foo": "bar",
			},
		},
		{
			name:           "remove all",
			valuesToRemove: []string{"foo"},
			valuesToKeep:   []string{},
			values: map[string]any{
				"foo":     "bar",
				"foo/bar": "baz",
			},
			expected: map[string]any{},
		},
		{
			name:           "remove multiple",
			valuesToRemove: []string{"foo/bar", "foo/test", "foo/foo", "c/d"},
			valuesToKeep:   []string{},
			values: map[string]any{
				"foo/bar":  "baz",
				"foo/test": "test",
				"foo/foo":  "foo",
				"a/b":      100,
				"c/d":      0,
			},
			expected: map[string]any{
				"a/b": 100,
			},
		},
		{
			name:           "allowlist",
			valuesToRemove: []string{},
			valuesToKeep:   []string{"foo/bar"},
			values: map[string]any{
				"foo/bar": "baz",
				"a/b":     100,
				"c/d":     0,
			},
			expected: map[string]any{
				"foo/bar": "baz",
			},
		},
		{
			name:           "allowlist and blocklist",
			valuesToRemove: []string{"c/d"},
			valuesToKeep:   []string{"foo/bar", "a/b", "b/c"},
			values: map[string]any{
				"foo/bar": "baz",
				"a/b":     100,
				"b/c":     200,
				"c/d":     0,
			},
			expected: map[string]any{
				"foo/bar": "baz",
				"a/b":     100,
				"b/c":     200,
			},
		},
		{
			name: "remove input-frames",
			valuesToRemove: []string{
				"access-point-oper-data/ethernet-if-stats/input-frames",
			},
			valuesToKeep: []string{},
			values: map[string]any{
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/duplex":       1,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name":      "LAN1",
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-crc":    0,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-drops":  0,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-errors": 0,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-frames": 0,
			},
			expected: map[string]any{
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/duplex":       1,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/if-name":      "LAN1",
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-crc":    0,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-drops":  0,
				"/rfc7951:/Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/input-errors": 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := removeValues(test.valuesToRemove, test.valuesToKeep, test.values)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetYangPathMatchers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			"single",
			"foo",
			[]string{
				"",
				"foo",
			},
		},
		{
			"top level",
			"Cisco-IOS-XE-device-hardware-oper:device-hardware-data",
			[]string{
				"",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data",
				"device-hardware-data",
			},
		},
		{
			"one level",
			"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware",
			[]string{
				"",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data",
				"device-hardware-data/device-hardware",
				"device-hardware-data",
			},
		},
		{
			"two levels",
			"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
			[]string{
				"",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware/device-inventory",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data/device-hardware",
				"Cisco-IOS-XE-device-hardware-oper:device-hardware-data",
				"device-hardware-data/device-hardware/device-inventory",
				"device-hardware-data/device-hardware",
				"device-hardware-data",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := getYangPathMatchers(test.input)

			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}

func TestParseEnum(t *testing.T) {
	t.Parallel()

	type input struct {
		yangPath     string
		value        string
		enumMappings map[string]map[string]map[string]int
	}

	type expected struct {
		v     int
		match bool
		err   error
	}

	tests := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			"single",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status",
				"oper-state-up",
				map[string]map[string]map[string]int{
					"access-point-oper-data": {
						"oper-status": {
							"oper-state-up": 1,
						},
					},
				},
			},
			expected{
				1,
				true,
				nil,
			},
		},
		{
			"none",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status",
				"oper-state-down",
				map[string]map[string]map[string]int{},
			},
			expected{
				0,
				false,
				nil,
			},
		},
		{
			"most specific to least specific matching",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status",
				"oper-state-down",
				map[string]map[string]map[string]int{
					"access-point-oper-data/ethernet-if-stats": {
						"oper-status": {
							"oper-state-down": 0,
						},
					},
					"access-point-oper-data": {
						"oper-status": {
							"oper-state-down": 1,
						},
					},
				},
			},
			expected{
				0,
				true,
				nil,
			},
		},
		{
			"less specific",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status",
				"oper-state-up",
				map[string]map[string]map[string]int{
					"access-point-oper-data/ethernet-if-stats": {
						"oper-state": {
							"oper-state-down": 0,
						},
					},
					"access-point-oper-data": {
						"oper-status": {
							"oper-state-up": 1,
						},
					},
				},
			},
			expected{
				1,
				true,
				nil,
			},
		},
		{
			"error",
			input{
				"Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status",
				"oper-state-unknown",
				map[string]map[string]map[string]int{
					"access-point-oper-data/ethernet-if-stats": {
						"oper-status": {
							"oper-state-up":   1,
							"oper-state-down": 0,
						},
					},
				},
			},
			expected{
				0,
				true,
				fmt.Errorf(
					"unknown value oper-state-unknown for enum: Cisco-IOS-XE-wireless-access-point-oper:access-point-oper-data/ethernet-if-stats/oper-status: %w",
					ErrEnumUnknownValue,
				),
			},
		},
		{
			"openconfig",
			input{
				"system/state/power",
				"up",
				map[string]map[string]map[string]int{
					"system": {
						"power": {
							"up": 1,
						},
					},
				},
			},
			expected{
				1,
				true,
				nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			v, match, err := parseEnum(
				test.input.yangPath,
				test.input.value,
				test.input.enumMappings,
			)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.match, match)
			assert.Equal(t, test.expected.v, v)
		})
	}
}

func TestTagsToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			"none",
			map[string]string{},
			"",
		},
		{
			"single",
			map[string]string{
				"tag1": "foo",
			},
			"tag1=foo",
		},
		{
			"multiple",
			map[string]string{
				"tag1": "foo",
				"tag2": "bar",
			},
			"tag1=foo,tag2=bar",
		},
		{
			"sorted",
			map[string]string{
				"zzz": "last",
				"aaa": "first",
			},
			"aaa=first,zzz=last",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := tagsToString(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
