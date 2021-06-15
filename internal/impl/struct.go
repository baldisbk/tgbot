package impl

import "time"

type Achievement struct {
	Name        string
	Description string
	CheckTime   time.Time
	Limit       int
	Initial     int
	Ascend      bool
	Done        bool
}

// Reach limit
type LimitAchievement struct {
	Achievement
	Current int
}

// Reach limit and hold it for 'stike' days
type StrikeAchievement struct {
	Achievement
	Best   int
	Last   int
	Strike int
}
