package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: reprocess-audio <bucket/key> [<bucket/key> ...]")
		os.Exit(2)
	}

	cli, err := minio.New(os.Getenv("S3_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
		Secure: os.Getenv("S3_USE_SSL") == "true",
		Region: os.Getenv("S3_REGION"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "minio:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	for _, p := range os.Args[1:] {
		bucket, key, ok := strings.Cut(strings.TrimLeft(p, "/"), "/")
		if !ok {
			fmt.Fprintln(os.Stderr, "skip invalid path:", p)
			continue
		}
		if err := reprocess(ctx, cli, bucket, key); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", p, err)
			os.Exit(1)
		}
	}
}

func reprocess(ctx context.Context, cli *minio.Client, bucket, key string) error {
	in := filepath.Join("/tmp", "in-"+filepath.Base(key))
	out := filepath.Join("/tmp", "out-"+filepath.Base(key))
	defer os.Remove(in)
	defer os.Remove(out)

	if err := cli.FGetObject(ctx, bucket, key, in, minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if channels(in) <= 2 {
		fmt.Println("SKIP (already stereo)", bucket+"/"+key)
		return nil
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-v", "error", "-i", in, // #nosec G204,G702 -- fixed binary with temp-file argv, no shell expansion.
		"-map", "0:v:0", "-map", "0:a:0", "-c:v", "copy", "-c:a", "aac", "-b:a", "160k", "-ac", "2",
		"-movflags", "+faststart", "-f", "mp4", out)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}

	if _, err := cli.FPutObject(ctx, bucket, key, out, minio.PutObjectOptions{ContentType: "video/mp4"}); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	fmt.Println("OK (downmixed to stereo)", bucket+"/"+key)
	return nil
}

func channels(path string) int {
	var stdout bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0", // #nosec G204,G702 -- fixed binary with temp-file argv, no shell expansion.
		"-show_entries", "stream=channels", "-of", "csv=p=0", path)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return 99
	}
	n, _ := strconv.Atoi(strings.TrimSpace(stdout.String()))
	return n
}
