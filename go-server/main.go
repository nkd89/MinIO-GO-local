package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endpoint        = "localhost:3334"
	accessKeyID     string
	secretAccessKey string
	bucketName      = "files"
	useSSL          = false
	baseURL         string
	uploadToken     string
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ No .env file found")
	}
	accessKeyID = os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
	baseURL = os.Getenv("BASE_URL")
	uploadToken = os.Getenv("UPLOAD_TOKEN")
}

func main() {
	port := flag.String("port", "3333", "port for the HTTP server")
	flag.Parse()

	ctx := context.Background()

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln("❌ MinIO client init error:", err)
	}

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalln("❌ Bucket check error:", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			log.Fatalln("❌ Bucket creation error:", err)
		}
	}

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != uploadToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "No file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		hash := fmt.Sprintf("%x", md5.Sum([]byte(header.Filename+fmt.Sprint(os.Getpid()))))

		_, err = client.PutObject(ctx, bucketName, hash, file, -1, minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		})
		if err != nil {
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}

		link := fmt.Sprintf("%s/%s", baseURL, hash)
		w.Write([]byte(link))
	})

	http.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/files/")
		object, err := client.GetObject(ctx, bucketName, key, minio.GetObjectOptions{})
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		io.Copy(w, object)
	})

	addr := ":" + *port
	log.Println("🚀 Go server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
