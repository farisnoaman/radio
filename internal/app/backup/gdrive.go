package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveProvider struct {
	ServiceAccountJSON string
	FolderID           string
	service            *drive.Service
}

func NewGoogleDriveProvider(serviceAccountJSON, folderID string) (*GoogleDriveProvider, error) {
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountJSON), drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %v", err)
	}

	return &GoogleDriveProvider{
		ServiceAccountJSON: serviceAccountJSON,
		FolderID:           folderID,
		service:            srv,
	}, nil
}

func (p *GoogleDriveProvider) Upload(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	filename := filepath.Base(filePath)
	file := &drive.File{
		Name:    filename,
		Parents: []string{p.FolderID},
	}

	_, err = p.service.Files.Create(file).Media(f).Do()
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// Additional methods for listing/deleting/downloading could be added if needed for full integration
