package repository

import (
	"context"
	"network-scanner/internal/models"
)

type Repository interface {
	// Сохранить запрос на сканирование
	SaveScanRequest(ctx context.Context, req *models.ScanRequest) (int64, error)

	// Сохранить результаты сканирования
	SaveScanResults(ctx context.Context, results []*models.ScanResult) error // Изменено на указатель

	// Получить историю сканирований
	GetScanHistory(ctx context.Context) ([]*models.ScanResponse, error)

	// Получить результаты сканирования по ID запроса
	GetScanResults(ctx context.Context, requestID int64) (*models.ScanResponse, error)

	// Закрыть соединение с базой данных
	Close() error
}
