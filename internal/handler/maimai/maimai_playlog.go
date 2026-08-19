package maimai

import (
	"sync"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

type queuedPlaylog struct {
	playlog   model.UserPlaylog
	expiresAt time.Time
}

var playlogBacklog = struct {
	sync.Mutex
	items map[int64][]queuedPlaylog
}{items: make(map[int64][]queuedPlaylog)}

// queuePlaylog retains at most six score uploads for a card that has not sent
// its initial UpsertUserAll yet, matching AquaDX's first-play behaviour.
func queuePlaylog(userID int64, playlog model.UserPlaylog) {
	playlogBacklog.Lock()
	defer playlogBacklog.Unlock()
	prunePlaylogBacklogLocked(time.Now())
	entries := append(playlogBacklog.items[userID], queuedPlaylog{playlog: playlog, expiresAt: time.Now().Add(10 * time.Minute)})
	if len(entries) > 6 {
		entries = entries[len(entries)-6:]
	}
	playlogBacklog.items[userID] = entries
}

func drainPlaylogs(userID int64) []model.UserPlaylog {
	playlogBacklog.Lock()
	defer playlogBacklog.Unlock()
	prunePlaylogBacklogLocked(time.Now())
	entries := playlogBacklog.items[userID]
	delete(playlogBacklog.items, userID)
	out := make([]model.UserPlaylog, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.playlog)
	}
	return out
}

func prunePlaylogBacklogLocked(now time.Time) {
	for userID, entries := range playlogBacklog.items {
		kept := entries[:0]
		for _, entry := range entries {
			if entry.expiresAt.After(now) {
				kept = append(kept, entry)
			}
		}
		if len(kept) == 0 {
			delete(playlogBacklog.items, userID)
		} else {
			playlogBacklog.items[userID] = kept
		}
	}
}
