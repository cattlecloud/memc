// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"cmp"
	"io"
	"regexp"
	"slices"
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

SCAN:
	// parse the contents of the stats output line by line
	for scanner.Scan() {
		line := scanner.Text()

		switch line {
		case "END":
			break SCAN

		case "ERROR":
			return nil, ErrCommandIssue

		default:
			fields := strings.Fields(line)
			if len(fields) < 3 || fields[0] != "STAT" {
				continue
			}
			key := fields[1]
			value := fields[2]
			m[key] = value
		}
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

type SlabStatistics struct {
	ActiveSlabs   int     `json:"active_slabs"`
	TotalMalloced int     `json:"total_malloced"`
	Slabs         []*Slab `json:"slabs"`
}

type Slab struct {
	Class         int `json:"slab_class"`
	ChunkSize     int `json:"chunk_size"`
	ChunksPerPage int `json:"chunks_per_page"`
	TotalPages    int `json:"total_pages"`
	TotalChunks   int `json:"total_chunks"`
	UsedChunks    int `json:"used_chunks"`
	FreeChunks    int `json:"free_chunks"`
	FreeChunksEnd int `json:"free_chunks_end"`
	GetHits       int `json:"get_hits"`
	CmdSet        int `json:"cmd_set"`
	DeleteHits    int `json:"delete_hits"`
	IncrementHits int `json:"incr_hits"`
	DecrementHits int `json:"decr_hits"`
	CASHits       int `json:"cas_hits"`
	CASBadVal     int `json:"cas_badval"`
	TouchHits     int `json:"touch_hits"`
}

var (
	statsSlabRe = regexp.MustCompile(`STAT (\d+):(\S+)\s+(\d+)`)
)

func slabs(r io.Reader) (*SlabStatistics, error) {
	scanner := bufio.NewScanner(r)

	stats := &SlabStatistics{
		Slabs: make([]*Slab, 0, 4),
	}

	m := make(map[int]*Slab)

SCAN:
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case line == "END":
			break SCAN

		case line == "ERROR":
			return nil, ErrCommandIssue

		case strings.HasPrefix(line, "STAT active_slabs"):
			fields := strings.Fields(line)
			active := toInt(fields[2])
			stats.ActiveSlabs = active

		case strings.HasPrefix(line, "STAT total_malloced"):
			fields := strings.Fields(line)
			malloced := toInt(fields[2])
			stats.TotalMalloced = malloced

		default:
			fields := statsSlabRe.FindStringSubmatch(line)
			if len(fields) != 4 {
				continue
			}
			slabclass := toInt(fields[1])
			name := fields[2]
			value := toInt(fields[3])

			if _, exists := m[slabclass]; !exists {
				m[slabclass] = &Slab{Class: slabclass}
			}
			slab := m[slabclass]

			switch name {
			case "chunk_size":
				slab.ChunkSize = value
			case "chunks_per_page":
				slab.ChunksPerPage = value
			case "total_pages":
				slab.TotalPages = value
			case "total_chunks":
				slab.TotalChunks = value
			case "used_chunks":
				slab.UsedChunks = value
			case "free_chunks":
				slab.FreeChunks = value
			case "free_chunks_end":
				slab.FreeChunksEnd = value
			case "get_hits":
				slab.GetHits = value
			case "cmd_set":
				slab.CmdSet = value
			case "delete_hits":
				slab.DeleteHits = value
			case "incr_hits":
				slab.IncrementHits = value
			case "decr_hits":
				slab.DecrementHits = value
			case "cas_hits":
				slab.CASHits = value
			case "cas_badval":
				slab.CASBadVal = value
			case "touch_hits":
				slab.TouchHits = value
			}
		}
	}

	// ensure the scan was a success
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, v := range m {
		stats.Slabs = append(stats.Slabs, v)
	}

	// order the slab classes ascending
	slices.SortFunc(stats.Slabs, func(a, b *Slab) int {
		return cmp.Compare(a.Class, b.Class)
	})

	return stats, nil
}

var (
	statsItemsRe = regexp.MustCompile(`STAT items:(\d+):(\S+)\s+(\d+)`)
)

type ItemStatistics struct {
	Class               int `json:"slab_class"`
	Number              int `json:"number"`
	NumberHot           int `json:"number_hot"`
	NumberWarm          int `json:"number_warm"`
	NumberCold          int `json:"number_cold"`
	AgeHot              int `json:"age_hot"`
	AgeWarm             int `json:"age_warm"`
	Age                 int `json:"age"`
	MemRequested        int `json:"mem_requested"`
	Evicted             int `json:"evicted"`
	EvictedNonZero      int `json:"evicted_nonzero"`
	EvictedTime         int `json:"evicted_time"`
	OutOfMemory         int `json:"outofmemory"`
	TailRepairs         int `json:"tailrepairs"`
	Reclaimed           int `json:"reclaimed"`
	ExpiredUnfetched    int `json:"expired_unfetched"`
	EvictedUnfetched    int `json:"evicted_unfetched"`
	EvictedActive       int `json:"evicted_active"`
	CrawlerReclaimed    int `json:"crawler_reclaimed"`
	CrawlerItemsChecked int `json:"crawler_items_checked"`
	LRUTailReflocked    int `json:"lrutail_reflocked"`
	MovesToCold         int `json:"moves_to_cold"`
	MovesToWarm         int `json:"moves_to_warm"`
	MovesWithinLRU      int `json:"moves_within_lru"`
	DirectReclaims      int `json:"direct_reclaims"`
	HitsToHot           int `json:"hits_to_hot"`
	HitsToWarm          int `json:"hits_to_warm"`
	HitsToCold          int `json:"hits_to_cold"`
	HitsToTemp          int `json:"hits_to_temp"`
}

func items(r io.Reader) ([]*ItemStatistics, error) {
	scanner := bufio.NewScanner(r)
	m := make(map[int]*ItemStatistics, 4)

SCAN:
	for scanner.Scan() {
		line := scanner.Text()

		switch line {
		case "END":
			break SCAN

		case "ERROR":
			return nil, ErrCommandIssue

		default:
			fields := statsItemsRe.FindStringSubmatch(line)
			if len(fields) != 4 {
				continue
			}
			slabclass := toInt(fields[1])
			name := fields[2]
			value := toInt(fields[3])

			if _, exists := m[slabclass]; !exists {
				m[slabclass] = new(ItemStatistics)
			}
			slab := m[slabclass]

			switch name {
			case "number":
				slab.Number = value
			case "number_hot":
				slab.NumberHot = value
			case "number_warm":
				slab.NumberWarm = value
			case "number_cold":
				slab.NumberCold = value
			case "age_hot":
				slab.AgeHot = value
			case "age_warm":
				slab.AgeWarm = value
			case "age":
				slab.Age = value
			case "mem_requested":
				slab.MemRequested = value
			case "evicted":
				slab.Evicted = value
			case "evicted_nonzero":
				slab.EvictedNonZero = value
			case "evicted_time":
				slab.EvictedTime = value
			case "outofmemory":
				slab.OutOfMemory = value
			case "tailrepairs":
				slab.TailRepairs = value
			case "reclaimed":
				slab.Reclaimed = value
			case "expired_unfetched":
				slab.ExpiredUnfetched = value
			case "evicted_unfetched":
				slab.EvictedUnfetched = value
			case "evicted_active":
				slab.EvictedActive = value
			case "crawler_reclaimed":
				slab.CrawlerReclaimed = value
			case "crawler_items_checked":
				slab.CrawlerItemsChecked = value
			case "lrutail_reflocked":
				slab.LRUTailReflocked = value
			case "moves_to_cold":
				slab.MovesToCold = value
			case "moves_to_warm":
				slab.MovesToWarm = value
			case "moves_within_lru":
				slab.MovesWithinLRU = value
			case "direct_reclaims":
				slab.DirectReclaims = value
			case "hits_to_hot":
				slab.HitsToHot = value
			case "hits_to_warm":
				slab.HitsToWarm = value
			case "hits_to_cold":
				slab.HitsToCold = value
			case "hits_to_temp":
				slab.HitsToTemp = value
			}
		}
	}

	// ensure the scan was a success
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	results := make([]*ItemStatistics, 0, len(m))
	for slabclass, v := range m {
		v.Class = slabclass
		results = append(results, v)
	}

	// order by slab class ascending
	slices.SortFunc(results, func(a, b *ItemStatistics) int {
		return cmp.Compare(a.Class, b.Class)
	})

	return results, nil
}
