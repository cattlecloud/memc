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
