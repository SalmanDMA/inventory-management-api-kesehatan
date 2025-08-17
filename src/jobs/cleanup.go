package jobs

// import (
// 	"log"
// 	"time"

// 	"github.com/SalmanDMA/inventory-app/backend/src/configs"
// 	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
// 	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
// )

// func StartFileCleanupScheduler() {
// 	go func() {
// 		for {
// 			now := time.Now()
// 			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
// 			if now.After(nextRun) {
// 				nextRun = nextRun.Add(24 * time.Hour)
// 			}

// 			durationUntilNextRun := time.Until(nextRun)
// 			log.Printf("[Cleanup Job] Sleeping until next run at %s (%s from now)", nextRun.Format("15:04:05"), durationUntilNextRun)

// 			time.Sleep(durationUntilNextRun)

// 			log.Println("[Cleanup Job] Starting file cleanup")
// 			deleteExpiredFiles()
// 		}
// 	}()
// }

// func deleteExpiredFiles() {
// 	uploadRepo := repositories.NewUploadRepository(configs.DB)
// 	files, err := uploadRepo.GetFilesToDelete(time.Now())
// 	if err != nil {
// 		log.Println("Error fetching expired uploads:", err)
// 		return
// 	}

// 	for _, file := range files {
// 		err := helpers.DeleteLocalFileImmediate(file.ID.String())
// 		if err != nil {
// 			log.Printf("Error deleting file %s: %v", file.ID, err)
// 		}
// 	}
// }
