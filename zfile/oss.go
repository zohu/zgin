package zfile

import (
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"io"
)

type ossService struct {
	client *oss.Client
}

func newOssService() *ossService {
	cfg := oss.LoadDefaultConfig().
		WithRegion(opts.Region).
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.AccessSecret))
	return &ossService{
		client: oss.NewClient(cfg),
	}
}

func (s *ossService) upload(ctx context.Context, rs io.ReadSeeker, name string, progress Progress) error {
	req := oss.PutObjectRequest{
		Bucket: oss.Ptr(opts.Bucket),
		Key:    oss.Ptr(name),
		Body:   rs,
	}
	if progress != nil {
		req.ProgressFn = func(increment, transferred, total int64) {
			progress(increment, transferred, total)
		}
	}
	_, err := s.client.PutObject(ctx, &req)
	if err != nil {
		return err
	}
	return nil
}
func (s *ossService) delete(ctx context.Context, name string) (err error) {
	_, err = s.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(opts.Bucket),
		Key:    oss.Ptr(name),
	})
	return err
}
