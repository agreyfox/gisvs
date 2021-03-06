package file

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/agreyfox/gisvs"
	"github.com/rs/zerolog"
)

type ServeService struct {
	FileStorage gisvs.FileStorage

	UsageService       gisvs.UsageService
	ApplicationService gisvs.ApplicationService
	TransformService   gisvs.TransformService

	FullFileDir string

	Log zerolog.Logger
}

func (s *ServeService) Serve(ctx context.Context, u *url.URL, opts gisvs.TransformationOption) (*gisvs.FileBlob, error) {
	fileBlobID := getFileBlobID(u.Path)

	file, err := s.FileStorage.FileByFileBlobID(ctx, fileBlobID)
	if err != nil {
		return nil, err
	}

	app, err := s.ApplicationService.Application(ctx, file.ApplicationID)
	if err != nil {
		return nil, err
	}

	fileBlobStorage, err := s.ApplicationService.FileBlobStorage(app.StorageEngine, app.StorageAccessKey, app.StorageSecretKey, app.StorageRegion, app.StorageEndpoint)
	if err != nil {
		return nil, err
	}

	fileBlob, err := fileBlobStorage.FileBlob(ctx, app.StorageBucket, fileBlobID, s.FullFileDir)
	if err != nil {
		return nil, fmt.Errorf("could not get file blob %w", err)
	}

	if !shouldTransform(file, opts) {
		updatedUsages := &gisvs.UpdateUsage{
			ApplicationID: file.ApplicationID,
			Bandwidth:     fileBlob.Size,
		}

		if err := s.UsageService.Update(ctx, updatedUsages); err != nil {
			// should I fail the request
			s.Log.Error().Err(err).Msg("failed to update usage")
		}

		return fileBlob, nil
	}

	// we can close the original fileBlob since we will be transforming it and generating a new one.
	// the returned blob gets closed by the parent of this function that still needs the blob around.
	defer fileBlob.Close()

	transformedBlob, err := s.TransformService.Transform(ctx, file, fileBlob, opts)
	if err != nil {
		return nil, err
	}

	updatedUsages := &gisvs.UpdateUsage{
		ApplicationID: file.ApplicationID,
		Bandwidth:     transformedBlob.Size,
	}

	if err := s.UsageService.Update(ctx, updatedUsages); err != nil {
		// should I fail the request
		s.Log.Error().Err(err).Msg("failed to update usage")
	}

	return transformedBlob, nil
}

func shouldTransform(file *gisvs.File, opts gisvs.TransformationOption) bool {
	return opts.Width > 0 ||
		opts.Height > 0 ||
		opts.Format != file.Extension
}

func getFileBlobID(urlPath string) string {
	path := strings.TrimPrefix(urlPath, "/")
	return strings.Split(path, ".")[0]
}
