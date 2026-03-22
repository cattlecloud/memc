// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"strings"
	"testing"

	"github.com/shoenig/test/must"
)

func Test_stats(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(realStats)
	result, err := stats(input)
	must.NoError(t, err)

	// spot check a few values
	must.Eq(t, 714, result.Runtime.PID)
	must.Eq(t, 1024, result.Connections.Max)
}

func Test_stats_slabs(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(realSlabsStats)
	result, err := slabs(input)
	must.NoError(t, err)
	must.Eq(t, 5, result.ActiveSlabs)
	must.Eq(t, 5242880, result.TotalMalloced)
	must.SliceLen(t, 5, result.Slabs)
	must.Eq(t, 600, result.Slabs[0].ChunkSize)
	must.Eq(t, 1747, result.Slabs[0].TotalChunks)
	must.Eq(t, 1856, result.Slabs[4].ChunkSize)
	must.Eq(t, 564, result.Slabs[4].TotalChunks)
}

func Test_stats_items(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(realStatsItems)
	result, err := items(input)
	must.NoError(t, err)
	must.SliceLen(t, 4, result)
	must.Eq(t, 9, result[0].Class)
	must.Eq(t, 6, result[0].Number)
	must.Eq(t, 3356, result[0].MemRequested)
}

// echo "stats" | nc -U /tmp/mc.sock
const realStats = `
STAT pid 714
STAT uptime 2077665
STAT time 1769190296
STAT version 1.6.29
STAT libevent 2.1.12-stable
STAT pointer_size 64
STAT rusage_user 551.779188
STAT rusage_system 808.853255
STAT max_connections 1024
STAT curr_connections 1
STAT total_connections 13714
STAT rejected_connections 0
STAT connection_structures 5
STAT response_obj_oom 0
STAT response_obj_count 1
STAT response_obj_bytes 65536
STAT read_buf_count 8
STAT read_buf_bytes 131072
STAT read_buf_bytes_free 49152
STAT read_buf_oom 0
STAT reserved_fds 20
STAT cmd_get 15249
STAT cmd_set 12260
STAT cmd_flush 0
STAT cmd_touch 0
STAT cmd_meta 0
STAT get_hits 2918
STAT get_misses 12331
STAT get_expired 78
STAT get_flushed 0
STAT delete_misses 0
STAT delete_hits 0
STAT incr_misses 0
STAT incr_hits 0
STAT decr_misses 0
STAT decr_hits 0
STAT cas_misses 0
STAT cas_hits 0
STAT cas_badval 0
STAT touch_hits 0
STAT touch_misses 0
STAT store_too_large 0
STAT store_no_memory 0
STAT auth_cmds 0
STAT auth_errors 0
STAT bytes_read 21752597
STAT bytes_written 125490335
STAT limit_maxbytes 2147483648
STAT accepting_conns 1
STAT listen_disabled_num 0
STAT time_in_listen_disabled_us 0
STAT threads 4
STAT conn_yields 0
STAT hash_power_level 16
STAT hash_bytes 524288
STAT hash_is_expanding 0
STAT slab_reassign_rescues 0
STAT slab_reassign_chunk_rescues 0
STAT slab_reassign_evictions_nomem 0
STAT slab_reassign_inline_reclaim 0
STAT slab_reassign_busy_items 0
STAT slab_reassign_busy_deletes 0
STAT slab_reassign_running 0
STAT slabs_moved 0
STAT lru_crawler_running 0
STAT lru_crawler_starts 1989
STAT lru_maintainer_juggles 29456575
STAT malloc_fails 0
STAT log_worker_dropped 0
STAT log_worker_written 0
STAT log_watcher_skipped 0
STAT log_watcher_sent 0
STAT log_watchers 0
STAT unexpected_napi_ids 0
STAT round_robin_fallback 0
STAT bytes 51321
STAT curr_items 11
STAT total_items 12260
STAT slab_global_page_pool 0
STAT expired_unfetched 11528
STAT evicted_unfetched 0
STAT evicted_active 0
STAT evictions 0
STAT reclaimed 11942
STAT crawler_reclaimed 130
STAT crawler_items_checked 3792
STAT lrutail_reflocked 619
STAT moves_to_cold 13354
STAT moves_to_warm 1368
STAT moves_within_lru 146
STAT direct_reclaims 0
STAT lru_bumps_dropped 0
END
`

// echo "stats slabs" | nc -U /tmp/mc.sock
const realSlabsStats = `
STAT 9:chunk_size 600
STAT 9:chunks_per_page 1747
STAT 9:total_pages 1
STAT 9:total_chunks 1747
STAT 9:used_chunks 6
STAT 9:free_chunks 1741
STAT 9:free_chunks_end 0
STAT 9:get_hits 2
STAT 9:cmd_set 10
STAT 9:delete_hits 0
STAT 9:incr_hits 0
STAT 9:decr_hits 0
STAT 9:cas_hits 0
STAT 9:cas_badval 0
STAT 9:touch_hits 0
STAT 11:chunk_size 944
STAT 11:chunks_per_page 1110
STAT 11:total_pages 1
STAT 11:total_chunks 1110
STAT 11:used_chunks 1
STAT 11:free_chunks 1109
STAT 11:free_chunks_end 0
STAT 11:get_hits 0
STAT 11:cmd_set 1
STAT 11:delete_hits 0
STAT 11:incr_hits 0
STAT 11:decr_hits 0
STAT 11:cas_hits 0
STAT 11:cas_badval 0
STAT 11:touch_hits 0
STAT 12:chunk_size 1184
STAT 12:chunks_per_page 885
STAT 12:total_pages 1
STAT 12:total_chunks 885
STAT 12:used_chunks 29
STAT 12:free_chunks 856
STAT 12:free_chunks_end 0
STAT 12:get_hits 0
STAT 12:cmd_set 29
STAT 12:delete_hits 0
STAT 12:incr_hits 0
STAT 12:decr_hits 0
STAT 12:cas_hits 0
STAT 12:cas_badval 0
STAT 12:touch_hits 0
STAT 13:chunk_size 1480
STAT 13:chunks_per_page 708
STAT 13:total_pages 1
STAT 13:total_chunks 708
STAT 13:used_chunks 86
STAT 13:free_chunks 622
STAT 13:free_chunks_end 0
STAT 13:get_hits 0
STAT 13:cmd_set 86
STAT 13:delete_hits 0
STAT 13:incr_hits 0
STAT 13:decr_hits 0
STAT 13:cas_hits 0
STAT 13:cas_badval 0
STAT 13:touch_hits 0
STAT 14:chunk_size 1856
STAT 14:chunks_per_page 564
STAT 14:total_pages 1
STAT 14:total_chunks 564
STAT 14:used_chunks 40
STAT 14:free_chunks 524
STAT 14:free_chunks_end 0
STAT 14:get_hits 0
STAT 14:cmd_set 40
STAT 14:delete_hits 0
STAT 14:incr_hits 0
STAT 14:decr_hits 0
STAT 14:cas_hits 0
STAT 14:cas_badval 0
STAT 14:touch_hits 0
STAT active_slabs 5
STAT total_malloced 5242880
END
`

// echo "stats items" | nc fossil.lan 11211
const realStatsItems = `
STAT items:9:number 6
STAT items:9:number_hot 0
STAT items:9:number_warm 0
STAT items:9:number_cold 6
STAT items:9:age_hot 0
STAT items:9:age_warm 0
STAT items:9:age 9
STAT items:9:mem_requested 3356
STAT items:9:evicted 0
STAT items:9:evicted_nonzero 0
STAT items:9:evicted_time 0
STAT items:9:outofmemory 0
STAT items:9:tailrepairs 0
STAT items:9:reclaimed 0
STAT items:9:expired_unfetched 0
STAT items:9:evicted_unfetched 0
STAT items:9:evicted_active 0
STAT items:9:crawler_reclaimed 0
STAT items:9:crawler_items_checked 0
STAT items:9:lrutail_reflocked 0
STAT items:9:moves_to_cold 6
STAT items:9:moves_to_warm 0
STAT items:9:moves_within_lru 0
STAT items:9:direct_reclaims 0
STAT items:9:hits_to_hot 2
STAT items:9:hits_to_warm 0
STAT items:9:hits_to_cold 0
STAT items:9:hits_to_temp 0
STAT items:12:number 1
STAT items:12:number_hot 0
STAT items:12:number_warm 0
STAT items:12:number_cold 1
STAT items:12:age_hot 0
STAT items:12:age_warm 0
STAT items:12:age 6
STAT items:12:mem_requested 1143
STAT items:12:evicted 0
STAT items:12:evicted_nonzero 0
STAT items:12:evicted_time 0
STAT items:12:outofmemory 0
STAT items:12:tailrepairs 0
STAT items:12:reclaimed 0
STAT items:12:expired_unfetched 0
STAT items:12:evicted_unfetched 0
STAT items:12:evicted_active 0
STAT items:12:crawler_reclaimed 0
STAT items:12:crawler_items_checked 0
STAT items:12:lrutail_reflocked 0
STAT items:12:moves_to_cold 1
STAT items:12:moves_to_warm 0
STAT items:12:moves_within_lru 0
STAT items:12:direct_reclaims 0
STAT items:12:hits_to_hot 0
STAT items:12:hits_to_warm 0
STAT items:12:hits_to_cold 0
STAT items:12:hits_to_temp 0
STAT items:13:number 3
STAT items:13:number_hot 0
STAT items:13:number_warm 0
STAT items:13:number_cold 3
STAT items:13:age_hot 0
STAT items:13:age_warm 0
STAT items:13:age 10
STAT items:13:mem_requested 4224
STAT items:13:evicted 0
STAT items:13:evicted_nonzero 0
STAT items:13:evicted_time 0
STAT items:13:outofmemory 0
STAT items:13:tailrepairs 0
STAT items:13:reclaimed 0
STAT items:13:expired_unfetched 0
STAT items:13:evicted_unfetched 0
STAT items:13:evicted_active 0
STAT items:13:crawler_reclaimed 0
STAT items:13:crawler_items_checked 0
STAT items:13:lrutail_reflocked 0
STAT items:13:moves_to_cold 3
STAT items:13:moves_to_warm 0
STAT items:13:moves_within_lru 0
STAT items:13:direct_reclaims 0
STAT items:13:hits_to_hot 0
STAT items:13:hits_to_warm 0
STAT items:13:hits_to_cold 0
STAT items:13:hits_to_temp 0
STAT items:14:number 5
STAT items:14:number_hot 0
STAT items:14:number_warm 0
STAT items:14:number_cold 5
STAT items:14:age_hot 0
STAT items:14:age_warm 0
STAT items:14:age 9
STAT items:14:mem_requested 7673
STAT items:14:evicted 0
STAT items:14:evicted_nonzero 0
STAT items:14:evicted_time 0
STAT items:14:outofmemory 0
STAT items:14:tailrepairs 0
STAT items:14:reclaimed 0
STAT items:14:expired_unfetched 0
STAT items:14:evicted_unfetched 0
STAT items:14:evicted_active 0
STAT items:14:crawler_reclaimed 0
STAT items:14:crawler_items_checked 0
STAT items:14:lrutail_reflocked 0
STAT items:14:moves_to_cold 5
STAT items:14:moves_to_warm 0
STAT items:14:moves_within_lru 0
STAT items:14:direct_reclaims 0
STAT items:14:hits_to_hot 0
STAT items:14:hits_to_warm 0
STAT items:14:hits_to_cold 0
STAT items:14:hits_to_temp 0
END
`
