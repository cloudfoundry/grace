#!/usr/bin/env bash

source $HOME/.bashisms/s3_upload.bash

set -e

echo "Compiling for linux..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
tar -zcf grace.tar.gz grace
echo "Uploading..."
upload_to_s3 grace.tar.gz
echo "Cleaning up..."
rm grace
rm grace.tar.gz

echo "Compiling for busybox..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "busybox" .
echo "Constructing Dockerimage"
docker build -t="cfdiegodocker/grace" .
docker push cfdiegodocker/grace
echo "Cleaning up..."
rm grace
