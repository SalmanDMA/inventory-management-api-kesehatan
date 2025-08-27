package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
)

func StartConsignmentDueReminderScheduler() {
	go func() {
		for {
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
			if !now.Before(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			d := time.Until(nextRun)
			log.Printf("[ConsignmentDue] Sleep until %s (in %s)\n", nextRun.Format(time.RFC3339), d)
			time.Sleep(d)

			if err := runConsignmentDueReminder(); err != nil {
				log.Printf("[ConsignmentDue] ERROR: %v\n", err)
			}
		}
	}()
}

func runConsignmentDueReminder() error {
	itemRepo := repositories.NewItemRepository(configs.DB)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
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
		daysLeft := int(it.DueDate.Truncate(24*time.Hour).Sub(start) / (24 * time.Hour))

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
			when, it.DueDate.Format("02 Jan 2006"), it.Code, it.Stock)

		metadata := map[string]interface{}{
			"item_id":    it.ID.String(),
			"item_name":  it.Name,
			"code":       it.Code,
			"due_date":   it.DueDate.Format(time.RFC3339),
			"days_left":  daysLeft,
			"is_consignment": it.IsConsignment,
		}

		if err := helpers.SendNotificationAuto("consigment_item", title, msg, metadata); err != nil {
			log.Printf("[ConsignmentDue] failed to send notif for item %s: %v\n", it.ID, err)
		}
	}

	log.Printf("[ConsignmentDue] Sent reminders for %d items\n", len(items))
	return nil
}
