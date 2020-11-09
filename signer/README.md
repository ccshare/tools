## S3 v4 signer

#### presign
- s3cli
```
s3cli -e http://192.168.55.2:9000 --presign cat open/h0 
http://192.168.55.2:9000/open/h0?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=object_user1%2F20201030%2Fdefault%2Fs3%2Faws4_request&X-Amz-Date=20201030T080624Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=4ae0512da1671e965d19e264794af9813cbf9bb30dcd5703419006b3977bc948
```

- signer
```
./signer --presign -Xget /open/h0
header:
  host:192.168.55.2:9000
url:
  http://192.168.55.2:9000/open/h1?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=object_user1%2F20201109%2Fdefault%2Fs3%2Faws4_request&X-Amz-Date=20201109T111506Z&X-Amz-Expires=43200&X-Amz-SignedHeaders=host&X-Amz-Signature=cc1be2b122dbe356d7ce8684b0ad41215c33df897c4105bb6aea0742b5ec33b8

```

#### signer
- s3cli
```
s3cli -e http://192.168.55.2:9000 --debug cat open/h0 
---[ REQUEST POST-SIGN ]-----------------------------
GET /open/h0 HTTP/1.1
Host: 192.168.55.2:9000
User-Agent: aws-sdk-go/0.24.0 (go1.15.2; darwin; amd64)
Amz-Sdk-Invocation-Id: 98BE55FD-BECC-4ACE-9A8C-7B9D35FB7CAB
Amz-Sdk-Request: attempt=1; max=3
Authorization: AWS4-HMAC-SHA256 Credential=object_user1/20201109/default/s3/aws4_request, SignedHeaders=amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=0a5785d53dedde111f7ec32a905d266f94978b722315292f3900aeeff4f59892
X-Amz-Content-Sha256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
X-Amz-Date: 20201109T111037Z
```

- signer
```
./signer -t 20201109T111037Z \
        -H'host:192.168.55.2:9000' \
        -H'Amz-Sdk-Invocation-Id:98BE55FD-BECC-4ACE-9A8C-7B9D35FB7CAB' \
        -H'X-Amz-Content-Sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' \
        -H'Amz-Sdk-Request:attempt=1; max=3' \
        /open/h0
header:
  Amz-Sdk-Invocation-Id:98BE55FD-BECC-4ACE-9A8C-7B9D35FB7CAB
  X-Amz-Content-Sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  Amz-Sdk-Request:attempt=1; max=3
  X-Amz-Date:20201109T111037Z
  Authorization:AWS4-HMAC-SHA256 Credential=object_user1/20201109/default/s3/aws4_request, SignedHeaders=amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=93f75717210a961bca66fc68dd00b20ab472695eef6a1b9d92b0a722842f0ae5
  Host:192.168.55.2:9000
url:
  http://192.168.55.2:9000/open/h1

```
