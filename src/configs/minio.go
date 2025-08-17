package configs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Minio *minio.Client

func InitMinio() {
	host := os.Getenv("MINIO_ENDPOINT")        
	port := os.Getenv("MINIO_PORT")            
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	ak := os.Getenv("MINIO_ACCESS_KEY")
	sk := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET_NAME")     

	if host == "" || port == "" {
		log.Fatal("MINIO_ENDPOINT or MINIO_PORT is empty")
	}
	if bucket == "" {
		log.Fatal("MINIO_BUCKET_NAME is empty")
	}

	endpoint := fmt.Sprintf("%s:%s", host, port)

	cl, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ak, sk, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("minio init error: %v", err)
	}
	Minio = cl

	ctx := context.Background()

	exists, err := Minio.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatalf("minio bucket check error: %v", err)
	}
	if !exists {
		if err := Minio.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			exists2, err2 := Minio.BucketExists(ctx, bucket)
			if err2 != nil || !exists2 {
				log.Fatalf("minio make bucket error: %v", err)
			}
		}
	}
}
