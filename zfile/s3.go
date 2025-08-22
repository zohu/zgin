package zfile

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/zohu/zgin/zlog"
)

type s3Service struct {
	client *s3.Client
}

func news3Service() *s3Service {
	cred := credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.AccessSecret, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(cred),
		config.WithRegion(opts.Region),
		config.WithBaseEndpoint(opts.Endpoint),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(so *retry.StandardOptions) {
				so.MaxAttempts = opts.MaxRetry
			})
		}),
	)
	if err != nil {
		zlog.Fatalf("初始化s3失败, %s", err.Error())
	}
	return &s3Service{
		client: s3.NewFromConfig(cfg, func(options *s3.Options) {

		}),
	}
}

func (s *s3Service) upload(ctx context.Context, rs io.ReadSeeker, name string, progress Progress) error {
	req := s3.PutObjectInput{
		Bucket: aws.String(opts.Bucket),
		Key:    aws.String(name),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   rs,
	}
	_, err := s.client.PutObject(ctx, &req)
	if err != nil {
		return err
	}
	return nil
}
func (s *s3Service) delete(ctx context.Context, name string) (err error) {
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(opts.Bucket),
		Key:    aws.String(name),
	})
	return err
}
