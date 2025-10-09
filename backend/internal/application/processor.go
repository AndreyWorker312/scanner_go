package application

import (
	"backend/domain/models"
	"backend/internal/application/services"
	rabbitmq "backend/internal/infrastructure/messaging"
)

// App представляет основное приложение с чистой архитектурой
type App struct {
	requestService   *services.RequestService
	responseService  *services.ResponseService
	publisherService *services.PublisherService
	historyService   *services.HistoryService
}

// NewApp создает новый экземпляр приложения
func NewApp(publisher *rabbitmq.RPCScannerPublisher, repo services.RepositoryInterface) *App {
	// Создаем сервисы
	historyService := services.NewHistoryService(repo)
	requestService := services.NewRequestService()
	responseService := services.NewResponseService(historyService)
	publisherService := services.NewPublisherService(publisher)

	// Устанавливаем callback для обработки ответов
	publisherService.SetResponseCallback(func(response *models.Response) {
		responseService.ProcessResponse(response)
	})

	return &App{
		requestService:   requestService,
		responseService:  responseService,
		publisherService: publisherService,
		historyService:   historyService,
	}
}

// ProcessRequest обрабатывает входящие запросы
func (a *App) ProcessRequest(req *models.Request) *models.Response {
	// Обрабатываем запрос
	response := a.requestService.ProcessRequest(req)

	// Если это валидный запрос, публикуем его
	if response.TaskID != "error" && response.TaskID != "unknown" {
		switch req.ScannerService {
		case "nmap_service":
			// Кэшируем запрос для последующего использования
			if nmapReq, ok := response.Result.(models.NmapTcpUdpRequest); ok {
				a.historyService.CacheRequest(nmapReq.TaskID, nmapReq)
			} else if nmapReq, ok := response.Result.(models.NmapOsDetectionRequest); ok {
				a.historyService.CacheRequest(nmapReq.TaskID, nmapReq)
			} else if nmapReq, ok := response.Result.(models.NmapHostDiscoveryRequest); ok {
				a.historyService.CacheRequest(nmapReq.TaskID, nmapReq)
			}
			return a.publisherService.PublishNmapRequest(response.Result)

		case "arp_service":
			if arpReq, ok := response.Result.(models.ARPRequest); ok {
				a.historyService.CacheRequest(arpReq.TaskID, arpReq)
				return a.publisherService.PublishARPRequest(arpReq)
			}

		case "icmp_service":
			if icmpReq, ok := response.Result.(models.ICMPRequest); ok {
				a.historyService.CacheRequest(icmpReq.TaskID, icmpReq)
				return a.publisherService.PublishICMPRequest(icmpReq)
			}
		}
	}

	return response
}

// ProcessResponse обрабатывает ответы от сканеров
func (a *App) ProcessResponse(response *models.Response) {
	a.responseService.ProcessResponse(response)
}

// PublishNmapRequest публикует Nmap запрос (для совместимости)
func (a *App) PublishNmapRequest(req interface{}) *models.Response {
	return a.publisherService.PublishNmapRequest(req)
}

// Методы для работы с историей (делегируются в historyService)
func (a *App) GetARPHistory(limit int) ([]models.ARPHistoryRecord, error) {
	return a.historyService.GetRepo().GetARPHistory(limit)
}

func (a *App) GetICMPHistory(limit int) ([]models.ICMPHistoryRecord, error) {
	return a.historyService.GetRepo().GetICMPHistory(limit)
}

func (a *App) GetNmapTcpUdpHistory(limit int) ([]models.NmapTcpUdpHistoryRecord, error) {
	return a.historyService.GetRepo().GetNmapTcpUdpHistory(limit)
}

func (a *App) GetNmapOsDetectionHistory(limit int) ([]models.NmapOsDetectionHistoryRecord, error) {
	return a.historyService.GetRepo().GetNmapOsDetectionHistory(limit)
}

func (a *App) GetNmapHostDiscoveryHistory(limit int) ([]models.NmapHostDiscoveryHistoryRecord, error) {
	return a.historyService.GetRepo().GetNmapHostDiscoveryHistory(limit)
}

func (a *App) DeleteARPHistory() error {
	return a.historyService.GetRepo().DeleteARPHistory()
}

func (a *App) DeleteICMPHistory() error {
	return a.historyService.GetRepo().DeleteICMPHistory()
}

func (a *App) DeleteNmapTcpUdpHistory() error {
	return a.historyService.GetRepo().DeleteNmapTcpUdpHistory()
}

func (a *App) DeleteNmapOsDetectionHistory() error {
	return a.historyService.GetRepo().DeleteNmapOsDetectionHistory()
}

func (a *App) DeleteNmapHostDiscoveryHistory() error {
	return a.historyService.GetRepo().DeleteNmapHostDiscoveryHistory()
}
