// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type Statistics struct {
	Runtime struct {
		PID         int    `json:"pid"`
		Uptime      int    `json:"uptime"`
		Time        int    `json:"time"`
		Version     string `json:"version"`
		LibEvent    string `json:"libevent"`
		PointerSize int    `json:"pointer_size"`
		Threads     int    `json:"threads"`
	}

	Resources struct {
		RUsageUser   float64 `json:"rusage_user"`
		RUsageSystem float64 `json:"rusage_system"`
	}

	Connections struct {
		Max        int `json:"max_connections"`
		Current    int `json:"curr_connections"`
		Total      int `json:"total_connections"`
		Rejected   int `json:"rejected_connections"`
		Structures int `json:"connection_structures"`
	}

	Commands struct {
		Get   int `json:"cmd_get"`
		Set   int `json:"cmd_set"`
		Flush int `json:"cmd_flush"`
		Touch int `json:"cmd_touch"`
		Meta  int `json:"cmd_meta"`

		Hit struct {
			Get       int `json:"get_hits"`
			Delete    int `json:"delete_hits"`
			Increment int `json:"incr_hits"`
			Decrement int `json:"decr_hits"`
			Touch     int `json:"touch_hits"`
			CAS       int `json:"cas_hits"`
		}

		Miss struct {
			Get       int `json:"get_misses"`
			Delete    int `json:"delete_misses"`
			Increment int `json:"incr_misses"`
			Decrement int `json:"decr_misses"`
			Touch     int `json:"touch_misses"`
			CAS       int `json:"cas_misses"`
		}

		Failure struct {
			GetExpired  int `json:"get_expired"`
			GetFlushed  int `json:"get_flushed"`
			CASBadValue int `json:"cas_badval"`
		}
	}

	Items struct {
		Bytes   int `json:"bytes"`
		Current int `json:"curr_items"`
		Total   int `json:"total_items"`
	}
}

func stats(r io.Reader) (*Statistics, error) {
	scanner := bufio.NewScanner(r)
	m := make(map[string]string)

	// parse the contents of the stats output line by line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "END" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "STAT" {
			continue
		}

		// skip fields[0], is STAT
		key := fields[1]
		value := fields[2]
		m[key] = value
	}

	// make sure the scanner was successful
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	s := new(Statistics)

	// map Runtime
	s.Runtime.PID = toInt(m["pid"])
	s.Runtime.Uptime = toInt(m["uptime"])
	s.Runtime.Time = toInt(m["time"])
	s.Runtime.Version = m["version"]
	s.Runtime.LibEvent = m["libevent"]
	s.Runtime.PointerSize = toInt(m["pointer_size"])
	s.Runtime.Threads = toInt(m["threads"])

	// map Resources
	s.Resources.RUsageUser = toFloat64(m["rusage_user"])
	s.Resources.RUsageSystem = toFloat64(m["rusage_system"])

	// map Connections
	s.Connections.Max = toInt(m["max_connections"])
	s.Connections.Current = toInt(m["curr_connections"])
	s.Connections.Total = toInt(m["total_connections"])
	s.Connections.Rejected = toInt(m["rejected_connections"])
	s.Connections.Structures = toInt(m["connection_structures"])

	// map Commands
	s.Commands.Get = toInt(m["cmd_get"])
	s.Commands.Set = toInt(m["cmd_set"])
	s.Commands.Flush = toInt(m["cmd_flush"])
	s.Commands.Touch = toInt(m["cmd_touch"])
	s.Commands.Meta = toInt(m["cmd_meta"])

	// map Hits
	s.Commands.Hit.Get = toInt(m["get_hits"])
	s.Commands.Hit.Delete = toInt(m["delete_hits"])
	s.Commands.Hit.Increment = toInt(m["incr_hits"])
	s.Commands.Hit.Decrement = toInt(m["decr_hits"])
	s.Commands.Hit.Touch = toInt(m["touch_hits"])
	s.Commands.Hit.CAS = toInt(m["cas_hits"])

	// map Misses
	s.Commands.Miss.Get = toInt(m["get_misses"])
	s.Commands.Miss.Delete = toInt(m["delete_misses"])
	s.Commands.Miss.Increment = toInt(m["incr_misses"])
	s.Commands.Miss.Decrement = toInt(m["decr_misses"])
	s.Commands.Miss.Touch = toInt(m["touch_misses"])
	s.Commands.Miss.CAS = toInt(m["cas_misses"])

	// map Failures
	s.Commands.Failure.GetExpired = toInt(m["get_expired"])
	s.Commands.Failure.GetFlushed = toInt(m["get_flushed"])
	s.Commands.Failure.CASBadValue = toInt(m["cas_badval"])

	// map Items
	s.Items.Bytes = toInt(m["bytes"])
	s.Items.Current = toInt(m["curr_items"])
	s.Items.Total = toInt(m["total_items"])

	return s, nil
}

func toInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func toFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
