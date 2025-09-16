package services

import "backend/domain/models"

type Scanner interface {
	GetType() models.ScanType
	Scan(request interface{}) (interface{}, error)
}

type ARPScanner interface {
	Scanner
	ScanARP(request *models.ARPRequest) (*models.ARPResponse, error)
}

type ICMPScanner interface {
	Scanner
	ScanICMP(request *models.ICMPRequest) (*models.ICMPResponse, error)
}

type NmapScanner interface {
	Scanner
	ScanTcpUdp(request *models.NmapTcpUdpRequest) (*models.NmapTcpUdpResponse, error)
	ScanOsDetection(request *models.NmapOsDetectionRequest) (*models.NmapOsDetectionResponse, error)
	ScanHostDiscovery(request *models.NmapHostDiscoveryRequest) (*models.NmapHostDiscoveryResponse, error)
}