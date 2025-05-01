package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type ServiceAccountKeyJSON struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

type GCSClient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
}

type IGCSClient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

func NewGCSClient(ServiceAccountKeyJSON ServiceAccountKeyJSON, BucketName string) IGCSClient {
	return &GCSClient{
		ServiceAccountKeyJSON: ServiceAccountKeyJSON,
		BucketName:            BucketName,
	}
}

func (g *GCSClient) createClient(ctx context.Context) (*storage.Client, error) {
	reqBodyBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("Error encoding service account key: %v", err)
		return nil, err
	}

	jsonBytes := reqBodyBytes.Bytes()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonBytes))
	if err != nil {
		logrus.Errorf("Error creating storage client: %v", err)
		return nil, err
	}

	return client, nil
}

func (g *GCSClient) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds = 60
	)

	client, err := g.createClient(ctx)
	if err != nil {
		logrus.Errorf("Error creating storage client: %v", err)
		return "", err
	}

	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("Error closing storage client: %v", err)
			return
		}
	}(client)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	defer cancel()

	bucket := client.Bucket(g.BucketName)
	object := bucket.Object(filename)
	buffer := bytes.NewBuffer(data)

	writer := object.NewWriter(ctx)
	writer.ChunkSize = 0

	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Errorf("Error copying data to writer: %v", err)
		return "", err
	}

	err = writer.Close()
	if err != nil {
		logrus.Errorf("failed to close: %v", err)
		return "", err
	}

	_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("Error updating object: %v", err)
		return "", err
	}

	fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, filename)

	return fileURL, nil
}
