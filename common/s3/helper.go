package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/http"
	"strings"
	"system-conf/common"
	"time"
)

func NewMinioClient(opts *MinioOptions) (mc *minio.Client, err error) {
	addrs := strings.Split(opts.Addr, ",")
	addr := addrs[0]
	var transport http.RoundTripper = nil
	if opts.SkipVerify && opts.Secure {
		transport = common.SkipVerify
	}
	mc, err = minio.New(addr, &minio.Options{
		Creds:        credentials.NewStaticV4(opts.Access, opts.Secret, ""),
		Secure:       opts.Secure,
		Transport:    transport,
		Region:       opts.Region,
		BucketLookup: 0,
		CustomMD5:    nil,
		CustomSHA256: nil,
	})

	return
}
func GetTimeoutContext(timeout time.Duration) context.Context {
	var ctx context.Context
	if timeout > 0 {
		ctx, _ = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, _ = context.WithCancel(context.Background())
	}
	return ctx
}
func List(mc *minio.Client, bucketName, prefix string, recursive bool, cb func(info minio.ObjectInfo)) {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	}
	for x := range mc.ListObjects(context.Background(), bucketName, opts) {

		if cb != nil {
			cb(x)
		} else {
			fmt.Println(x)
		}
	}
}
func EnsureBucket(mc *minio.Client, bucketName string, opts minio.MakeBucketOptions, timeout time.Duration) (err error) {
	exists, e := mc.BucketExists(GetTimeoutContext(timeout), bucketName)
	if e != nil {
		err = e
		return
	}
	if !exists {
		err = mc.MakeBucket(GetTimeoutContext(timeout), bucketName, opts)
	}
	return
}
func Read(mc *minio.Client, bucketName, key string, options minio.GetObjectOptions) (reader io.Reader, err error) {
	var obj *minio.Object
	obj, err = mc.GetObject(context.Background(), bucketName, key, options)
	if err != nil {
		return
	}
	//var info minio.ObjectInfo
	//info, err = obj.Stat()
	//if err !=nil{
	//	return
	//}
	//buf := make([]byte, info.Size)
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, obj)
	if err == nil {
		reader = buf
	}
	return
}
func ReadData(mc *minio.Client, bucketName, key string, options minio.GetObjectOptions, timeout time.Duration) (data []byte, err error) {
	var obj *minio.Object
	obj, err = mc.GetObject(GetTimeoutContext(timeout), bucketName, key, options)
	if err != nil {
		return
	}
	data, err = io.ReadAll(obj)
	if err == nil && len(data) > 3 {
		if data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
			data = data[3:]
		}
	}
	return
}

func WriteData(mc *minio.Client, bucketName, key string, data []byte, options minio.PutObjectOptions, timeout time.Duration) (sz int, err error) {
	var uploadInfo minio.UploadInfo
	uploadInfo, err = mc.PutObject(GetTimeoutContext(timeout), bucketName, key, bytes.NewBuffer(data), int64(len(data)), options)
	if err == nil {
		sz = int(uploadInfo.Size)
	}
	return
}
