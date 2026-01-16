// Copyright (c) 2026 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package driller

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// Driller navigates JSON using a flexible dot path supporting arrays
func Driller(jsonData string, path string) gjson.Result {
	parts := strings.Split(path, ".")
	current := gjson.Parse(jsonData)

	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(\[(\d|\*)?\])?$`)

	for _, p := range parts {
		matches := re.FindStringSubmatch(p)
		if len(matches) == 0 {
			return gjson.Result{} // Invalid path segment
		}

		key := matches[1]

		// matches[2] is the [], which we can throw away.

		index := -1
		if matches[3] != "" {
			// Array index specified
			i, err := strconv.Atoi(matches[3])
			if err != nil {
				return gjson.Result{}
			}
			index = i
		}

		val := current.Get(key)
		if val.IsArray() {
			// If index is specified, use it; otherwise default to [0]
			arr := val.Array()
			switch {
			case index == -1:
				if len(arr) == 1 {
					val = arr[0]
				}
				// Otherwise do nothing. We'll dump the whole list.
			case index >= 0 && index < len(arr):
				val = arr[index]
			default:
				return gjson.Result{}
			}
		}

		current = val
	}

	return current
}
