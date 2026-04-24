package main

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/derekparker/trie/v3"
	"github.com/itchyny/gojq"
	"github.com/openconfig/gnmic/pkg/formatters"
)

var (
	yangOriginRegexp        = regexp.MustCompile(`^\/?(?:[^:\/]+:\/|\/)`)
	yangPathNamespaceRegexp = regexp.MustCompile(`^([^:\/]+:)?`)
	yangPathRootRegexp      = regexp.MustCompile(`^([^\/]+)`)
	tagBaseRegexp           = regexp.MustCompile(`^[^_]+_`)
	timestampRegexp         = regexp.MustCompile(
		`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[\+\-]\d{2}:\d{2})$`,
	)
	jsonArrayRegexp = regexp.MustCompile(`(?P<name>.+)\.(?P<idx>[0-9]+)$`)
)

// Strip the origin (e.g openconfig:/) from a YANG path.
func stripYangPathOrigin(yangPath string) string {
	return yangOriginRegexp.ReplaceAllString(strings.TrimSuffix(yangPath, "/"), "")
}

// Strip the origin and namespace (e.g openconfig:/openconfig:) from a YANG path.
func stripYangPathNamespace(yangPath string) string {
	return yangPathNamespaceRegexp.ReplaceAllString(stripYangPathOrigin(yangPath), "")
}

// Strip anything that is not the root (e.g openconfig:system) from a YANG path.
func getYangPathRoot(yangPath string) string {
	return yangPathRootRegexp.FindString(stripYangPathOrigin(yangPath))
}

// Strip YANG model information from a tag.
func stripTags(tags map[string]string) map[string]string {
	newTags := map[string]string{}

	if tags == nil {
		return newTags
	}

	for k, v := range tags {
		base := tagBaseRegexp.ReplaceAllString(k, "")
		newTags[base] = v
	}

	return newTags
}

// Add a new tag to a set of tags, returns an error if new tag already exists with a different value.
func addTag(tags map[string]string, tag string, value string) error {
	value = strings.TrimSpace(value)
	if cur, exists := tags[tag]; !exists || value == cur || cur == "" {
		tags[tag] = value
		return nil
	}

	return fmt.Errorf(
		"tag %v exists with a different value (%v != %v): %w",
		tag,
		tags[tag],
		value,
		ErrTagExists,
	)
}

// Strip YANG model information from a value.
func stripValues(values map[string]any, metricPrefix string) map[string]any {
	newValues := map[string]any{}

	if values == nil {
		return newValues
	}

	for k, v := range values {
		prefix := []string{metricPrefix}
		base := yangPathNamespaceRegexp.ReplaceAllString(k, strings.Join(prefix, "/")+"/")
		newValues[base] = v
	}

	return newValues
}

// Apply a compiled JQ expression to a set of messages and separately return the events
// that match the expression as well as the events that don't match the expression.
func matchEvents(
	events []*formatters.EventMsg,
	condition *gojq.Code,
) ([]*formatters.EventMsg, []*formatters.EventMsg, error) {
	if condition == nil {
		return events, []*formatters.EventMsg{}, nil
	}

	matchedEvents := []*formatters.EventMsg{}
	unmatchedEvents := []*formatters.EventMsg{}
	errs := []error{}

	for _, e := range events {
		match, err := formatters.CheckCondition(condition, e)
		if err != nil {
			errs = append(errs, err)
		} else {
			if match {
				matchedEvents = append(matchedEvents, e)
			} else {
				unmatchedEvents = append(unmatchedEvents, e)
			}
		}
	}

	return matchedEvents, unmatchedEvents, errors.Join(errs...)
}

// Split json values (e.g .0 ... .n) into separate events.
func splitEventsByJSONArray(events []*formatters.EventMsg) []*formatters.EventMsg {
	splitEvents := []*formatters.EventMsg{}

	for _, e := range events {
		eventsByJSONArrayIdx := map[string]*formatters.EventMsg{}
		for k, v := range e.Values {
			jsonArrayIdx := ""
			jsonArrayBase := ""
			pathElems := strings.Split(k, "/")
			for i, p := range pathElems {
				match := jsonArrayRegexp.FindStringSubmatch(p)
				if len(match) == 3 {
					jsonArrayBase = strings.Join(pathElems[:i+1], "/")
					pathElems[i] = match[jsonArrayRegexp.SubexpIndex("name")]
					jsonArrayIdx = match[jsonArrayRegexp.SubexpIndex("idx")]
					break
				}
			}
			if jsonArrayIdx != "" && jsonArrayBase != "" {
				if eventsByJSONArrayIdx[jsonArrayBase] == nil {
					tags := map[string]string{
						"json_idx": jsonArrayIdx,
					}
					maps.Copy(tags, e.Tags)
					eventsByJSONArrayIdx[jsonArrayBase] = &formatters.EventMsg{
						Name:      e.Name,
						Timestamp: e.Timestamp,
						Tags:      tags,
						Values:    map[string]any{},
					}
				}

				strippedName := strings.Join(pathElems, "/")
				eventsByJSONArrayIdx[jsonArrayBase].Values[strippedName] = v
				delete(e.Values, k)
			}
		}

		for _, e := range eventsByJSONArrayIdx {
			splitEvents = append(splitEvents, e)
		}

		if len(e.Values) >= 1 || len(e.Deletes) >= 1 {
			splitEvents = append(splitEvents, e)
		}
	}

	return splitEvents
}

// Group events with the same tags together.
func groupEventsByTag(events []*formatters.EventMsg) map[string]*formatters.EventMsg {
	eventsByTags := map[string]*formatters.EventMsg{}
	for _, e := range events {
		tagString := fmt.Sprintf("%s", e.Tags)
		if eventsByTags[tagString] == nil {
			eventsByTags[tagString] = &formatters.EventMsg{
				Name:      e.Name,
				Timestamp: e.Timestamp,
				Tags:      e.Tags,
			}
		}
		if len(e.Values) >= 1 {
			if eventsByTags[tagString].Values == nil {
				eventsByTags[tagString].Values = map[string]any{}
			}
			maps.Copy(eventsByTags[tagString].Values, e.Values)
		}
		if len(e.Deletes) >= 1 {
			if eventsByTags[tagString].Deletes == nil {
				eventsByTags[tagString].Deletes = []string{}
			}
			eventsByTags[tagString].Deletes = append(eventsByTags[tagString].Deletes, e.Deletes...)
		}
	}
	return eventsByTags
}

// Comparator function for string length.
func compareLength(a, b string) int {
	return cmp.Compare(len(a), len(b))
}

// Convert nested values in an event into a tree structure.
func eventToTree(event *formatters.EventMsg) (*trie.Trie[*formatters.EventMsg], error) {
	eventTree := trie.New[*formatters.EventMsg]()

	errs := []error{}

	for _, k := range slices.SortedFunc(maps.Keys(event.Values), compareLength) {
		v := event.Values[k]
		k = stripYangPathOrigin(k)
		dir := filepath.Dir(k)

		if node, exists := eventTree.Find(dir); exists {
			if cur, exists := node.Val().Values[k]; !exists || v == cur || cur == "" {
				node.Val().Values[k] = v
			} else {
				errs = append(errs, fmt.Errorf(
					"value %v exists with a different value (%v != %v): %w",
					k,
					node.Val().Values[k],
					v,
					ErrTagExists,
				))
			}
		} else {
			tags := map[string]string{}
			maps.Copy(tags, event.Tags)

			eventTree.Add(dir, &formatters.EventMsg{
				Name:      dir,
				Timestamp: event.Timestamp,
				Tags:      tags,
				Values:    map[string]any{k: v},
			})
		}
	}

	slices.SortFunc(event.Deletes, compareLength)
	for _, del := range event.Deletes {
		del = stripYangPathOrigin(del)

		tags := map[string]string{}
		maps.Copy(tags, event.Tags)

		eventTree.Add(del, &formatters.EventMsg{
			Name:      del,
			Timestamp: event.Timestamp,
			Tags:      tags,
			Deletes:   []string{del},
		})
	}

	return eventTree, errors.Join(errs...)
}

// Apply a number of rename operations to a set of tags.
func renameTags(
	yangPath string,
	tags map[string]string,
	tagRenames map[string]map[string]string,
) (map[string]string, error) {
	yangPath = stripYangPathOrigin(yangPath)

	newTags := map[string]string{}
	maps.Copy(newTags, tags)

	errs := []error{}

	for _, matcher := range getYangPathMatchers(yangPath) {
		if tagRenames, exists := tagRenames[matcher]; exists {
			for oldName, newName := range tagRenames {
				if v, exists := newTags[oldName]; exists {
					if err := addTag(newTags, newName, v); err != nil {
						errs = append(
							errs,
							fmt.Errorf(
								"error renaming %v to %v: %w",
								oldName,
								newName,
								err,
							),
						)
					}
					delete(newTags, oldName)
				}
			}
		}
	}

	return newTags, errors.Join(errs...)
}

// Merge values together that share the same YANG path.
func mergeValues(values map[string]any) map[string]map[string]any {
	merged := map[string]map[string]any{}

	for k, v := range values {
		dir := filepath.Dir(k)
		base := filepath.Base(k)

		if _, exists := merged[dir]; !exists {
			merged[dir] = map[string]any{}
		}

		merged[dir][base] = v
	}

	return merged
}

// Apply an allow/block list to a set of values.
func removeValues(
	valuesToRemove []string,
	valuesToKeep []string,
	values map[string]any,
) map[string]any {
	newValues := map[string]any{}

	maps.Copy(newValues, values)

	maps.DeleteFunc(newValues, func(k string, _ any) bool {
		for _, matcher := range getYangPathMatchers(k) {
			if slices.Contains(valuesToRemove, matcher) {
				return true
			}
		}
		return false
	})

	if len(valuesToKeep) >= 1 {
		maps.DeleteFunc(newValues, func(k string, _ any) bool {
			for _, matcher := range getYangPathMatchers(k) {
				if slices.Contains(valuesToKeep, matcher) {
					return false
				}
			}
			return true
		})
	}

	return newValues
}

// Parse a timestamp string into a time.Time.
func parseTimestamp(s string) (time.Time, bool, error) {
	if timestampRegexp.MatchString(s) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("error parsing timestamp: %w", err)
		}

		return t, true, nil
	}

	return time.Time{}, false, nil
}

// Get all possible matchers that match a given YANG path.
func getYangPathMatchers(yangPath string) []string {
	yangPath = stripYangPathOrigin(yangPath)

	matchers := []string{}

	currDir := yangPath
	for {
		matchers = append(matchers, currDir)

		strippedCurrDir := stripYangPathNamespace(currDir)
		if currDir != strippedCurrDir {
			matchers = append(matchers, strippedCurrDir)
		}

		currDir = filepath.Dir(currDir)

		if currDir == "." || currDir == "/" {
			break
		}
	}

	matchers = append(matchers, "")

	return matchers
}

// Parse an enum string value to an integer.
func parseEnum(
	yangPath string,
	value string,
	enumMappings map[string]map[string]map[string]int,
) (int, bool, error) {
	yangPath = stripYangPathOrigin(yangPath)
	dir := filepath.Dir(yangPath)
	base := filepath.Base(yangPath)

	for _, matcher := range getYangPathMatchers(dir) {
		if m, exists := enumMappings[matcher]; exists {
			if e, exists := m[base]; exists {
				if x, exists := e[value]; exists {
					return x, true, nil
				}
				return 0, true, fmt.Errorf(
					"unknown value %s for enum: %s: %w",
					value,
					yangPath,
					ErrEnumUnknownValue,
				)
			}
		}
	}

	return 0, false, nil
}

// Convert a set of tags to a string.
func tagsToString(tags map[string]string) string {
	entries := make([]string, 0, len(tags))
	for _, k := range slices.Sorted(maps.Keys(tags)) {
		entries = append(entries, fmt.Sprintf("%s=%s", k, tags[k]))
	}
	return strings.Join(entries, ",")
}
