package mock

import (
	"context"
	"mime/multipart"

	"github.com/agreyfox/gisvs"
)

type UploadService struct {
}

func (s *UploadService) Upload(ctx context.Context, r *multipart.Reader) (*gisvs.File, error) {
	return nil, nil
}

func (s *UploadService) ChunkUpload(ctx context.Context, r *multipart.Reader) error {
	return nil
}

func (s *UploadService) CompleteChunkUpload(ctx context.Context, appID, uploadID, filename string) (*gisvs.File, error) {
	return nil, nil
}
