package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/gobuffalo/envy"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const prefix = ""

var (
	storageClient *storage.Client
	bucket        string
	credsJSON     string

	attrList = []string{"Name", "ContentType", "Size"}
)

func main() {
	envy.Load()
	bucket = envy.Get("GCLOUD_BUCKET_NAME", "attract")
	credsJSON = envy.Get("GCLOUD_STORAGE_CREDS", "")
	ctx := context.Background()

	var err error
	creds, err := google.CredentialsFromJSON(ctx, []byte(credsJSON), storage.ScopeFullControl)
	if err != nil {
		log.Fatal("Error parsing credential from JSON ", err)
	}

	storageClient, err = storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatal("Error creating new storage client ", err)
	}

	bkt := storageClient.Bucket(bucket)
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		log.Fatal("error reading bucket")
	}
	fmt.Printf("bucket %s, created at %s, is located in %s with storage class %s\n",
		attrs.Name, attrs.Created, attrs.Location, attrs.StorageClass)
	query := &storage.Query{Prefix: prefix}
	query.SetAttrSelection(attrList)
	it := bkt.Objects(ctx, query)
	var totalSaved int64
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			fmt.Println("Done")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if !strings.HasPrefix(attrs.ContentType, "image") {
			continue
		}
		if attrs.Size < 1000000 {
			continue
		}
		// fmt.Printf("%s\t%s\t%d\t\n", attrs.Name, attrs.ContentType, attrs.Size)
		obj := bkt.Object(attrs.Name)
		r, err := obj.NewReader(ctx)
		if err != nil {
			fmt.Printf("failed to read obj: %s -- %s", attrs.Name, err.Error())
			continue
		}
		crushedFile, err := crushFile(r)
		r.Close()
		if err != nil {
			fmt.Printf("failed to crush image: %s -- %s", attrs.Name, err.Error())
			continue
		}
		w := obj.NewWriter(ctx)
		_, err = io.Copy(w, crushedFile)
		w.Close()
		if err != nil {
			fmt.Printf("failed to write crushed image: %s -- %s", attrs.Name, err.Error())
			continue
		}
		newAttrs, err := obj.Attrs(ctx)
		if err != nil {
			fmt.Printf("failed to get new attrs: %s -- %s", attrs.Name, err.Error())
			continue
		}
		saved := attrs.Size - newAttrs.Size
		totalSaved += saved
		crushPercentage := float64(saved) / float64(attrs.Size) * 100
		fmt.Printf("%s\t%.2f%%\t%d\t\n\n", newAttrs.Name, crushPercentage, totalSaved)
	}

	fmt.Printf("Total saved: %d\n", totalSaved)
}

// SaveFileToStorage persists the file to Storage
func SaveFileToStorage(ctx context.Context, filename string, contents io.Reader) error {
	obj := objHandle(ctx, filename)
	sw := obj.NewWriter(ctx)
	if _, err := io.Copy(sw, contents); err != nil {
		err = fmt.Errorf("Could not write file: %w", err)
		return err
	}

	if err := sw.Close(); err != nil {
		err = fmt.Errorf("Could not put file: %w", err)
		return err
	}

	return nil
}

func objHandle(ctx context.Context, filename string) *storage.ObjectHandle {
	return storageClient.Bucket(bucket).Object(filename)
}
func crushFile(file io.Reader) (*bytes.Buffer, error) {
	img, err := imaging.Decode(file)
	if err != nil {
		err = fmt.Errorf("failed to decode image for crushing: %w", err)
		return nil, err
	}
	resized := imaging.Resize(img, 800, 0, imaging.Lanczos)
	buf := &bytes.Buffer{}
	err = imaging.Encode(buf, resized, imaging.PNG)
	if err != nil {
		err = fmt.Errorf("failed to encode crushed file: %w", err)
	}
	return buf, err
}
