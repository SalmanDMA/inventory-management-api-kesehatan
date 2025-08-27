package jobs

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/minio/minio-go/v7"
)

func StartDatabaseBackupScheduler(loc *time.Location) {
	go func() {
		for {
			now := time.Now().In(loc)
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, loc)
			for nextRun.Weekday() != time.Sunday {
				nextRun = nextRun.Add(24 * time.Hour)
			}
			if !now.Before(nextRun) {
				nextRun = nextRun.Add(7 * 24 * time.Hour)
			}

			d := time.Until(nextRun)
			log.Printf("[DBBackup] Sleep until %s (in %s)\n", nextRun.Format(time.RFC3339), d)
			time.Sleep(d)

			if err := runDatabaseBackup(); err != nil {
				log.Printf("[DBBackup] ERROR: %v\n", err)
			}
		}
	}()
}

func runDatabaseBackup() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
					sslmode = "disable"
	}

	args := []string{
					"-h", host,
					"-p", port,
					"-U", user,
					"-d", dbname,
					"-F", "p",
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+pass)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
					return fmt.Errorf("pg_dump failed: %v, output: %s", err, out.String())
	}

	timestamp := time.Now().Format("2006-01-02_150405")
	sqlFile := fmt.Sprintf("backup_%s.sql", timestamp)
	zipFile := fmt.Sprintf("backup_%s.zip", timestamp)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	w, err := zipWriter.Create(sqlFile)
	if err != nil {
					return fmt.Errorf("zip create: %w", err)
	}
	if _, err := io.Copy(w, &out); err != nil {
					return fmt.Errorf("zip copy: %w", err)
	}
	zipWriter.Close()

	objectKey := fmt.Sprintf("backups/%s/%s", time.Now().Format("2006-01"), zipFile)
	_, err = configs.Minio.PutObject(
					context.Background(),
					os.Getenv("MINIO_BUCKET_NAME"),
					objectKey,
					&zipBuf,
					int64(zipBuf.Len()),
					minio.PutObjectOptions{ContentType: "application/zip"},
	)
	if err != nil {
					return fmt.Errorf("failed to upload backup: %w", err)
	}

	log.Printf("[DBBackup] ✅ Backup uploaded: %s (%d bytes)", objectKey, zipBuf.Len())
	return nil
}
