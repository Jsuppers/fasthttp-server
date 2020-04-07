[![Build Status](https://travis-ci.com/Jsuppers/fasthttp-server.svg?branch=master)](https://travis-ci.com/Jsuppers/fasthttp-server)
[![Coverage Status](https://coveralls.io/repos/github/Jsuppers/fasthttp-server/badge.svg?branch=master&service=github)](https://coveralls.io/github/Jsuppers/fasthttp-server?branch=master)

# fasthttp-server
fasthttp-server is a service which receives messages from fasthttp-client (https://github.com/Jsuppers/fasthttp-client), once received the service:
* extracts the clientID, which is used to indicate where this message will be saved
* formats the message in http://ndjson.org/
* compresses the message in gzip format
* streams this data either to s3 storage or a Azure blob

## Configuration
to assign which storage to stream to set the STORAGE_TYPE environment variable e.g. to stream to azure:
```
export STORAGE_TYPE="azure"
```
note: if not set it will default to s3 storage
## how to run in docker
```
git clone https://github.com/Jsuppers/fasthttp-server.git
docker build -t fasthttp-server .
```
#### stream to S3 storage
```
docker run --rm -it -p 8080:8080 --env AWS_BUCKET="bucket" --env AWS_REGION="region" --env AWS_ACCESS_KEY="key" --env AWS_ACCESS_SECRET="secret" fasthttp-server
```
#### stream to Azure blob
```
docker run --rm -it -p 8080:8080 --env STORAGE_TYPE="azure" --env AZURE_STORAGE_ACCOUNT="account" --env AZURE_STORAGE_ACCESS_KEY="key" fasthttp-server
```
