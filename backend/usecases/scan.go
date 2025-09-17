package usecases

import (
	"context"
	"fmt"
	"time"

	"backend/domain/models"
	"backend/domain/repository"
	"backend/domain/services"
)

type ScanUseCases struct {
	scanRepo    repository.ScanRepository
	arpScanner  services.ARPScanner
	icmpScanner services.ICMPScanner
	nmapScanner services.NmapScanner
}

func NewScanUseCases(
	scanRepo repository.ScanRepository,
	arpScanner services.ARPScanner,
	icmpScanner services.ICMPScanner,
	nmapScanner services.NmapScanner,
) *ScanUseCases {
	return &ScanUseCases{
		scanRepo:    scanRepo,
		arpScanner:  arpScanner,
		icmpScanner: icmpScanner,
		nmapScanner: nmapScanner,
	}
}

func (uc *ScanUseCases) StartScan(ctx context.Context, scanType models.ScanType, request interface{}) (*models.ScanRequest, error) {
	scanRequest := &models.ScanRequest{
		TaskID:    generateTaskID(scanType),
		Type:      scanType,
		Request:   request,
		Status:    models.ScanStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.scanRepo.CreateScan(scanRequest); err != nil {
		return nil, err
	}

	go uc.executeScan(scanRequest)

	return scanRequest, nil
}

func (uc *ScanUseCases) executeScan(scanRequest *models.ScanRequest) {
	uc.scanRepo.UpdateScanStatus(scanRequest.TaskID, models.ScanStatusRunning)

	startTime := time.Now()
	var response interface{}
	var err error

	switch scanRequest.Type {
	case models.ScanTypeARP:
		if arpRequest, ok := scanRequest.Request.(*models.ARPRequest); ok {
			response, err = uc.arpScanner.ScanARP(arpRequest)
		}
	case models.ScanTypeICMP:
		if icmpRequest, ok := scanRequest.Request.(*models.ICMPRequest); ok {
			response, err = uc.icmpScanner.ScanICMP(icmpRequest)
		}
	case models.ScanTypeNMAP:
		response, err = uc.handleNmapScan(scanRequest.Request)
	}

	duration := time.Since(startTime).Milliseconds()

	// Сохраняем результат
	scanResponse := &models.ScanResponse{
		TaskID:    scanRequest.TaskID,
		Type:      scanRequest.Type,
		Response:  response,
		Error:     getErrorString(err),
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	if err != nil {
		scanResponse.Status = "failed"
		uc.scanRepo.UpdateScanStatus(scanRequest.TaskID, models.ScanStatusFailed)
	} else {
		scanResponse.Status = "completed"
		uc.scanRepo.UpdateScanStatus(scanRequest.TaskID, models.ScanStatusCompleted)
	}

	uc.scanRepo.SaveScanResponse(scanResponse)
}

func (uc *ScanUseCases) handleNmapScan(request interface{}) (interface{}, error) {
	// Определяем тип Nmap сканирования по структуре запроса
	switch req := request.(type) {
	case *models.NmapTcpUdpRequest:
		return uc.nmapScanner.ScanTcpUdp(req)
	case *models.NmapOsDetectionRequest:
		return uc.nmapScanner.ScanOsDetection(req)
	case *models.NmapHostDiscoveryRequest:
		return uc.nmapScanner.ScanHostDiscovery(req)
	default:
		return nil, fmt.Errorf("unknown nmap request type")
	}
}

func generateTaskID(scanType models.ScanType) string {
	return string(scanType) + "-" + time.Now().Format("20060102150405")
}

func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
