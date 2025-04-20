// backend/service/report_service.go
package service

import (
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"time"
)

// ReportService bietet Funktionen für Lagerberichte
type ReportService struct {
	articleRepo     *repository.ArticleRepository
	transactionRepo *repository.TransactionRepository
	supplierRepo    *repository.SupplierRepository
}

// NewReportService erstellt einen neuen ReportService
func NewReportService() *ReportService {
	return &ReportService{
		articleRepo:     repository.NewArticleRepository(),
		transactionRepo: repository.NewTransactionRepository(),
		supplierRepo:    repository.NewSupplierRepository(),
	}
}

// GenerateStockValueReport erstellt einen Bericht über den Lagerwert
func (s *ReportService) GenerateStockValueReport() (map[string]interface{}, error) {
	// Gesamtwert des Lagerbestands
	totalStockValue, err := s.articleRepo.CalculateTotalStockValue()
	if err != nil {
		return nil, err
	}

	// Lagerwert nach Kategorien
	categorySummary, err := s.articleRepo.GetCategorySummary()
	if err != nil {
		categorySummary = map[string][]interface{}{
			"labels": {"Keine Kategorie"},
			"counts": {0},
			"values": {0.0},
		}
	}

	// Top-Artikel nach Wert
	// Dies würde eine neue Methode im ArticleRepository erfordern
	// topArticlesByValue, err := s.articleRepo.FindTopByValue(10)
	// if err != nil {
	//     topArticlesByValue = []*model.Article{}
	// }

	return map[string]interface{}{
		"totalStockValue": totalStockValue,
		"categorySummary": categorySummary,
		//"topArticlesByValue": topArticlesByValue,
		"generatedAt": time.Now(),
	}, nil
}

// GenerateInventoryTurnoverReport erstellt einen Bericht über die Lagerumschlagshäufigkeit
func (s *ReportService) GenerateInventoryTurnoverReport(startDate, endDate time.Time) (map[string]interface{}, error) {
	// Dieser Bericht würde zusätzliche Daten im TransactionRepository erfordern
	// und ist hier nur als Platzhalter implementiert

	return map[string]interface{}{
		"period": map[string]string{
			"start": startDate.Format("02.01.2006"),
			"end":   endDate.Format("02.01.2006"),
		},
		"turnoverRate": 0.0, // Beispielwert
		"generatedAt":  time.Now(),
	}, nil
}

// GenerateLowStockReport erstellt einen Bericht über Artikel unter Mindestbestand
func (s *ReportService) GenerateLowStockReport() ([]*model.Article, error) {
	return s.articleRepo.FindLowStock(0) // 0 = keine Begrenzung
}

// GenerateStockMovementReport erstellt einen Bericht über Bestandsbewegungen
func (s *ReportService) GenerateStockMovementReport(startDate, endDate time.Time) (map[string]interface{}, error) {
	// Dieser Bericht würde zusätzliche Daten im TransactionRepository erfordern
	// und ist hier nur als Platzhalter implementiert

	return map[string]interface{}{
		"period": map[string]string{
			"start": startDate.Format("02.01.2006"),
			"end":   endDate.Format("02.01.2006"),
		},
		"stockInTotal":  0.0, // Beispielwert
		"stockOutTotal": 0.0, // Beispielwert
		"generatedAt":   time.Now(),
	}, nil
}
