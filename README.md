# Google Cloud storage crawler

Crawls all or a prefix of a GCloud storage bucket and performs
modifications on the objects. The current implementation scales
images down to 800px wide. But modifying for other purposes is
straightforward.

## Usage

```bash
$ export GCLOUD_BUCKET_NAME=my-bucket-name
$ export GCLOUD_STORAGE_CREDS='{"json_of": "storage_creds"}'
$ go build
$ ./storage-crawler
```
