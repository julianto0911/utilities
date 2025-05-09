package utilities

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UploadS3File(ctx context.Context, basePath, imagePath string) (string, error) {
	args := m.Called(ctx, basePath, imagePath)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) RemoveS3File(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func NewStorage(s3 S3) Storage {
	return &storage{s3: s3}
}

func GetS3Config() S3Config {
	cfg := S3Config{
		Bucket:          EnvString("S3_BUCKET"),
		Path:            EnvString("S3_PATH"),
		Region:          EnvString("S3_REGION"),
		AccessKeyID:     EnvString("S3_ACCESS_KEY_ID"),
		SecretAccessKey: EnvString("S3_SECRET_ACCESS_KEY"),
		Token:           EnvString("S3_TOKEN"),
		Endpoint:        EnvString("S3_ENDPOINT"),
		S3TenantID:      EnvString("S3_TENANT_ID"),
		DoStream:        EnvBool("S3_DO_STREAM"),
		Acl:             EnvString("S3_ACL"),
	}

	return cfg
}

type Storage interface {
	UploadS3File(ctx context.Context, basePath, imagePath string) (string, error)
	RemoveS3File(path string) error
}

type storage struct {
	s3 S3
}

func (u *storage) UploadS3File(ctx context.Context, basePath, imagePath string) (string, error) {
	errHandle := func(err error) (string, error) {
		RemoveFile(imagePath)
		return "", err
	}

	_, stream, err := FileStream(imagePath)
	if err != nil {
		return errHandle(fmt.Errorf("create filestream > %w", err))
	}

	_, err = u.s3.CreateFolder(basePath)
	if err != nil {
		return errHandle(fmt.Errorf("create folder > %w", err))
	}

	ext := filepath.Ext(imagePath)
	filename := NewUUID() + ext

	//prepare upload
	uploader, err := u.s3.InitUploader()
	if err != nil {
		return errHandle(fmt.Errorf("init uploader > %w", err))
	}

	path := basePath + filename
	uploadParams := &s3manager.UploadInput{
		Bucket:      aws.String(u.s3.BucketName()),
		Key:         aws.String(path),
		Body:        stream,
		ContentType: aws.String(ext),
		ACL:         aws.String(u.s3.AclValue()),
	}

	_, err = uploader.UploadWithContext(ctx, uploadParams)
	if err != nil {
		return errHandle(fmt.Errorf("upload file > %w", err))
	}

	RemoveFile(imagePath)

	//return full public path
	fullPath := u.s3.EndpointValue() + "/" + u.s3.BucketName() + path

	return fullPath, nil
}

func (u *storage) RemoveS3File(path string) error {
	if u.s3 == nil {
		return errors.New("s3 instance is nil")
	}

	if path == "" {
		return errors.New("empty path provided")
	}

	//remove the public route prefix
	filePath := strings.Replace(path, u.s3.EndpointValue()+"/"+u.s3.BucketName(), "", 1)

	_, err := u.s3.DeleteImage(filePath)
	return err
}

func RemoveFolder(path string) {
	os.RemoveAll(path)
}

func RemoveFile(path string) {
	os.Remove(path)
}
