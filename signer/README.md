## S3 v4 signer

- presing
s3cli -e http://192.168.55.2:9000 --presign cat open/h0 
http://192.168.55.2:9000/open/h0?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=object_user1%2F20201030%2Fdefault%2Fs3%2Faws4_request&X-Amz-Date=20201030T080624Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=4ae0512da1671e965d19e264794af9813cbf9bb30dcd5703419006b3977bc948

./signer -s http://192.168.55.2:9000/open/h0
presign url:  http://192.168.55.2:9000/open/h0?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=object_user1%2F20201030%2Fdefault%2Fs3%2Faws4_request&X-Amz-Date=20201030T080624Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=445d32165e2a551cd09c0cde36834c87511ddb2f93cc66884d8446d9d36425bb
presign header:  map[host:[192.168.55.2:9000]]


- sign
s3cli -e http://192.168.55.2:9000 --debug cat open/h0 
2020/10/30 16:08:25 DEBUG: Request Amazon S3/GetObject Details:
---[ REQUEST POST-SIGN ]-----------------------------
GET /open/h0 HTTP/1.1
Host: 192.168.55.2:9000
User-Agent: aws-sdk-go/0.23.0 (go1.14.4; darwin; amd64)
Amz-Sdk-Invocation-Id: C090A090-094B-4F2C-8C57-D2B4763911AC
Amz-Sdk-Request: attempt=1; max=3
Authorization: AWS4-HMAC-SHA256 Credential=object_user1/20201030/default/s3/aws4_request, SignedHeaders=amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=889fe6993dd2c21aba4568e6b8c7cd89f9cb832e059f7ee9ababbdd5cdb99472
X-Amz-Content-Sha256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
X-Amz-Date: 20201030T080825Z
Accept-Encoding: gzip

