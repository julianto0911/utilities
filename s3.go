package utilities

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func NewS3(cfg S3Config) S3 {
	s3 := cloudStorage{}
	s3.AccessKeyID = cfg.AccessKeyID
	s3.SecretAccessKey = cfg.SecretAccessKey
	s3.Token = cfg.Token
	s3.Region = cfg.Region
	s3.Endpoint = cfg.Endpoint
	s3.Bucket = cfg.Bucket
	s3.S3TenantID = cfg.S3TenantID
	s3.Path = cfg.Path
	s3.Acl = cfg.Acl

	return &s3
}

type S3 interface {
	InitUploader() (*s3manager.Uploader, error)
	CreateFolder(path string) (*s3.PutObjectOutput, error)
	DeleteImage(path string) (*s3.DeleteObjectOutput, error)
	GetFileList(ctx context.Context, path string) ([]string, error)
	DownloadFile(file *os.File, path string) (int64, error)
	BucketName() string
	RegionValue() string
	AccessKeyIDValue() string
	SecretAccessKeyValue() string
	TokenValue() string
	EndpointValue() string
	S3TenantIDValue() string
	PathValue() string
	AclValue() string
}

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Token           string
	Region          string
	Endpoint        string
	Bucket          string
	S3TenantID      string
	DoStream        bool
	Path            string
	Acl             string
}

type cloudStorage struct {
	AccessKeyID     string
	SecretAccessKey string
	Token           string
	Region          string
	Endpoint        string
	Bucket          string
	S3TenantID      string
	Path            string
	Acl             string
}

func (c *cloudStorage) AclValue() string {
	return c.Acl
}

func (c *cloudStorage) RegionValue() string {
	return c.Region
}

func (c *cloudStorage) AccessKeyIDValue() string {
	return c.AccessKeyID
}

func (c *cloudStorage) SecretAccessKeyValue() string {
	return c.SecretAccessKey
}

func (c *cloudStorage) TokenValue() string {
	return c.Token
}

func (c *cloudStorage) EndpointValue() string {
	return c.Endpoint
}

func (c *cloudStorage) BucketValue() string {
	return c.Bucket
}

func (c *cloudStorage) S3TenantIDValue() string {
	return c.S3TenantID
}

func (c *cloudStorage) PathValue() string {
	return c.Path
}

func (c *cloudStorage) getSVC() (*s3.S3, error) {
	creds := credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, c.Token)
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}

	cfg := aws.NewConfig().WithRegion(c.Region).WithCredentials(creds).WithEndpoint(c.Endpoint).WithS3ForcePathStyle(true)
	mySession := session.Must(session.NewSession())
	return s3.New(mySession, cfg), nil
}

func (c *cloudStorage) BucketName() string {
	return c.Bucket
}

func (c *cloudStorage) initDownloader() (*s3manager.Downloader, error) {
	//get credentials
	creds := credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, c.Token)
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}

	//create session
	cfg := aws.NewConfig().WithRegion(c.Region).WithCredentials(creds).WithEndpoint(c.Endpoint).WithS3ForcePathStyle(true)
	s3Session, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(s3Session)

	return downloader, nil
}

func (c *cloudStorage) InitUploader() (*s3manager.Uploader, error) {
	//get credentials
	creds := credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, c.Token)
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}

	//create session
	cfg := aws.NewConfig().WithRegion(c.Region).WithCredentials(creds).WithEndpoint(c.Endpoint).WithS3ForcePathStyle(true)
	s3Session, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	//create uploader
	uploader := s3manager.NewUploader(s3Session)

	return uploader, nil
}

func (c *cloudStorage) CreateFolder(path string) (*s3.PutObjectOutput, error) {
	svc, err := c.getSVC()
	if err != nil {
		return nil, err
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(path),
		ACL:    &c.Acl,
	}

	resp, err := svc.PutObject(params)
	return resp, err
}

func (c *cloudStorage) DeleteImage(path string) (*s3.DeleteObjectOutput, error) {
	svc, err := c.getSVC()
	if err != nil {
		return nil, err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(path),
	}

	// use the iterator to delete the files.
	resp, err := svc.DeleteObject(input)
	return resp, err
}

func (c *cloudStorage) GetFileList(ctx context.Context, path string) ([]string, error) {
	//get credentials
	creds := credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, c.Token)
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}

	//create session
	cfg := aws.NewConfig().WithRegion(c.Region).WithCredentials(creds).WithEndpoint(c.Endpoint).WithS3ForcePathStyle(true)
	s3Session, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	s3client := s3.New(s3Session)

	s3Keys := make([]string, 0)

	if err := s3client.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(c.Bucket),
		Prefix: aws.String(path), // list files in the directory.
	}, func(o *s3.ListObjectsOutput, b bool) bool { // callback func to enable paging.
		for _, o := range o.Contents {
			s3Keys = append(s3Keys, *o.Key)
		}
		return true
	}); err != nil {
		return nil, err
	}

	return s3Keys, nil
}

func (c *cloudStorage) DownloadFile(file *os.File, path string) (int64, error) {
	downloader, err := c.initDownloader()
	if err != nil {
		return 0, err
	}

	byt, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(c.Bucket),
			Key:    aws.String(path),
		})
	if err != nil {
		return 0, err
	}

	return byt, err
}
