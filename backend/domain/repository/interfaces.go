package repository

import "backend/domain/models"


type ScanRepository interface {
	CreateScan(scan *models.ScanRequest) error
	GetScanByTaskID(taskID string) (*models.ScanRequest, error)
	UpdateScanStatus(taskID string, status models.ScanStatus) error
	SaveScanResponse(response *models.ScanResponse) error
	GetScanResponses(taskID string) ([]*models.ScanResponse, error)
}