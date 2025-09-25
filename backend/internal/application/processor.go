package application

import (
	"backend/domain/models"
	rabbitmq "backend/internal/infrastructure"
	"log"
)

type App struct {
	publisher *rabbitmq.RPCScannerPublisher
}

func NewApp(publisher *rabbitmq.RPCScannerPublisher) *App {
	return &App{
		publisher: publisher,
	}
}

func (a *App) ProcessRequest(req *models.Request) *models.Response {
	var resp *models.Response
	switch req.ScannerService {
	case "nmap_service":
		resp = a.nmapValidation(req.Options)
	case "arp_service":
		resp = a.arpValidation(req.Options)
	case "icmp_service":
		resp = a.icmpValidation(req.Options)
	default:
		log.Fatal("Failed to switching process request service")
	}
	return resp
}

func (a *App) nmapValidation(req any) *models.Response {
	nmapReq, ok := req.(models.NmapRequest)
	if !ok {
		log.Fatal("Failed to validate nmap scanner")
	}
	nmapResp, err := a.publisher.PublishNmap(nmapReq)
	if err != nil {
		log.Fatal("Failed to publish nmap task: ", err)
	}
	return nmapResp
}

func (a *App) arpValidation(req any) *models.Response {
	arpReq, ok := req.(models.ARPRequest)
	if !ok {
		log.Fatal("Failed to validate arp scanner")
	}
	arpResp, err := a.publisher.PublishArp(arpReq)
	if err != nil {
		log.Fatal("Failed to publish arp task: ", err)
	}
	return arpResp
}

func (a *App) icmpValidation(req any) *models.Response {
	icmpReq, ok := req.(models.ICMPRequest)
	if !ok {
		log.Fatal("Failed to validate icmp scanner")
	}
	icmpResp, err := a.publisher.PublishIcmp(icmpReq)
	if err != nil {
		log.Fatal("Failed to publish icmp task: ", err)
	}
	return icmpResp
}
