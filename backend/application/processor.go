package websocket

import (
	"backend/domain/models"
	"log"
)

func ProcessRequest(req *models.Request) *models.Response {
	var resp *models.Response
	switch req.ScannerService {
	case "nmap_service":
		resp = nmapValidaion(req.Options)
	case "arp_service":
		resp = arpValidaion(req.Options)
	case "icmp_service":
		resp = icmpValidaion(req.Options)
	default:
		log.Fatal("Faild to switching process request service")
	}
	return resp
}

func nmapValidaion(req any) *models.Response {
	nmapReq, ok := req.(models.NmapRequest)
	if !ok {
		log.Fatal("Faild to validate nmap scanner")
	}
	nmapResp := PublishNmap(nmapReq)
	return nmapResp
}

func arpValidaion(req any) *models.Response {
	arpReq, ok := req.(models.ARPRequest)
	if !ok {
		log.Fatal("Faild to validate arp scanner")
	}
	arpResp := PublishArp(arpReq)
	return arpResp
}

func icmpValidaion(req any) *models.Response {
	icmpReq, ok := req.(models.ICMPRequest)
	if !ok {
		log.Fatal("Faild to validate icmp scanner")
	}
	icmpResp := PublishIcmp(icmpReq)
	return icmpResp
}
