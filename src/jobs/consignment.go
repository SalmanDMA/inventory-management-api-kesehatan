package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
)

func StartConsignmentDueReminderScheduler(loc *time.Location) {
	go func() {
		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, loc)
			if !now.Before(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			d := time.Until(nextRun)
			log.Printf("[ConsignmentDue] Sleep until %s (in %s)\n", nextRun.Format(time.RFC3339), d)
			time.Sleep(d)

			if err := runConsignmentDueReminder(loc); err != nil {
				log.Printf("[ConsignmentDue] ERROR: %v\n", err)
			}
		}
	}()
}

func runConsignmentDueReminder(loc *time.Location) error {
	itemRepo := repositories.NewItemRepository(configs.DB)

	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 3).Add(24*time.Hour - time.Nanosecond)

	items, err := itemRepo.FindConsignmentDueBetween(nil, start, end)
	if err != nil {
		return fmt.Errorf("query items: %w", err)
	}
	if len(items) == 0 {
		log.Println("[ConsignmentDue] No items in window")
		return nil
	}

	for _, it := range items {
		if it.DueDate == nil {
			continue
		}
		daysLeft := int(it.DueDate.In(loc).Truncate(24*time.Hour).Sub(start) / (24 * time.Hour))

		title := fmt.Sprintf("Pengingat Konsinyasi: %s", it.Name)
		var when string
		switch daysLeft {
		case 0:
			when = "Jatuh tempo hari ini"
		case 1, 2, 3:
			when = fmt.Sprintf("Jatuh tempo %d hari lagi", daysLeft)
		default:
			when = "Mendekati jatuh tempo"
		}

		msg := fmt.Sprintf("%s (Due: %s). Kode: %s, Stock: %d.",
			when, it.DueDate.In(loc).Format("02 Jan 2006"), it.Code, it.Stock)

		metadata := map[string]interface{}{
			"item_id":       it.ID.String(),
			"item_name":     it.Name,
			"code":          it.Code,
			"due_date":      it.DueDate.In(loc).Format(time.RFC3339),
			"days_left":     daysLeft,
			"is_consignment": it.IsConsignment,
		}

		if err := helpers.SendNotificationAuto("consigment_item", title, msg, metadata); err != nil {
			log.Printf("[ConsignmentDue] failed to send notif for item %s: %v\n", it.ID, err)
		}
	}

	log.Printf("[ConsignmentDue] Sent reminders for %d items\n", len(items))
	return nil
}
