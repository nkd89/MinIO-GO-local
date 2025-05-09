package main

import (
	"context"
	"crypto/md5"
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
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	accessKeyID = os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey = os.Getenv("MINIO_SECRET_KEY")
	baseURL = os.Getenv("BASE_URL")
}

func main() {
	ctx := context.Background()

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	if exists, err := client.BucketExists(ctx, bucketName); err != nil {
		log.Fatalln(err)
	} else if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			log.Fatalln(err)
		}
	}

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
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

	log.Println("Go server on :3333")
	log.Fatal(http.ListenAndServe(":3333", nil))
}
