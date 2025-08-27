package jobs

import "time"

func StartAll(loc *time.Location) {
	StartConsignmentDueReminderScheduler(loc)
}
