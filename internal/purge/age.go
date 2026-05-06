package purge

import "time"

// DefaultRecentDays is the recency threshold for marking a Target as
// "Recent". Mole uses 7 days; we keep parity so users coming from Mole
// see consistent defaults.
const DefaultRecentDays = 7

// IsRecent reports whether modified is within recentDays days of now.
// Pass DefaultRecentDays to use the project default. A zero modified
// time is never considered recent so missing stat info errs on the side
// of "safe to purge".
func IsRecent(now, modified time.Time, recentDays int) bool {
	if modified.IsZero() {
		return false
	}
	if recentDays <= 0 {
		recentDays = DefaultRecentDays
	}
	return now.Sub(modified) < time.Duration(recentDays)*24*time.Hour
}
