// backend/service/stock_service.go
package service

import (
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StockService verwaltet die Lagerbestandsfunktionen
type StockService struct {
	articleRepo     *repository.ArticleRepository
	transactionRepo *repository.TransactionRepository
	activityRepo    *repository.ActivityRepository
}

// NewStockService erstellt einen neuen StockService
func NewStockService() *StockService {
	return &StockService{
		articleRepo:     repository.NewArticleRepository(),
		transactionRepo: repository.NewTransactionRepository(),
		activityRepo:    repository.NewActivityRepository(),
	}
}

// PerformStockAdjustment führt eine Bestandsanpassung durch
func (s *StockService) PerformStockAdjustment(
	articleID string,
	transactionType model.TransactionType,
	quantity float64,
	reason, reference, notes string,
	userID primitive.ObjectID,
	userName string,
) (*model.Transaction, error) {
	// Artikel abrufen
	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		return nil, fmt.Errorf("Artikel nicht gefunden: %v", err)
	}

	// Alten Bestand speichern
	oldStock := article.StockCurrent

	// Neuen Bestand berechnen basierend auf dem Transaktionstyp
	var newStock float64
	switch transactionType {
	case model.TransactionTypeStockIn:
		newStock = oldStock + quantity
	case model.TransactionTypeStockOut:
		newStock = oldStock - quantity
		// Prüfen, ob genügend Bestand vorhanden ist
		if newStock < 0 {
			return nil, fmt.Errorf("Nicht genügend Bestand vorhanden")
		}
	case model.TransactionTypeAdjust, model.TransactionTypeInventory:
		// Bei Anpassung/Inventur ist die Menge bereits der neue Bestand
		newStock = quantity
		// Anpassen der Menge für die Transaktion (Differenz zum alten Bestand)
		quantity = newStock - oldStock
	default:
		return nil, fmt.Errorf("Ungültiger Transaktionstyp")
	}

	// Neue Transaktion erstellen
	transaction := &model.Transaction{
		ID:          primitive.NewObjectID(),
		Type:        transactionType,
		ArticleID:   article.ID,
		ArticleName: article.ShortName,
		Quantity:    quantity,
		OldStock:    oldStock,
		NewStock:    newStock,
		UnitPrice:   article.PurchasePriceNet,
		Reason:      reason,
		Reference:   reference,
		UserID:      userID,
		UserName:    userName,
		Timestamp:   time.Now(),
		Notes:       notes,
	}

	// Transaktion speichern
	err = s.transactionRepo.Create(transaction)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Speichern der Transaktion: %v", err)
	}

	// Artikelbestand aktualisieren
	article.StockCurrent = newStock
	if transactionType == model.TransactionTypeInventory {
		article.LastStockTakeDate = time.Now()
	}

	err = s.articleRepo.Update(article)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Aktualisieren des Artikelbestands: %v", err)
	}

	// Aktivität loggen
	activityType := model.ActivityTypeStockAdjusted
	if transactionType == model.TransactionTypeInventory {
		activityType = model.ActivityTypeStockTaking
	}

	_, _ = s.activityRepo.LogActivity(
		activityType,
		userID,
		userName,
		article.ID,
		"article",
		article.ShortName,
		fmt.Sprintf("%s: %g %s (neu: %g)", transaction.GetDisplayType(), quantity, article.Unit, newStock),
		quantity,
	)

	return transaction, nil
}

// CheckLowStockArticles prüft, ob Artikel unter Mindestbestand sind
func (s *StockService) CheckLowStockArticles() ([]*model.Article, error) {
	return s.articleRepo.FindLowStock(0) // 0 = keine Begrenzung
}

// GetStockStatus gibt einen Statusbericht über den Lagerbestand zurück
func (s *StockService) GetStockStatus() (map[string]interface{}, error) {
	// Gesamtzahl der Artikel
	totalArticles, err := s.articleRepo.Count()
	if err != nil {
		return nil, err
	}

	// Artikel unter Mindestbestand
	lowStockArticles, err := s.articleRepo.FindLowStock(0)
	if err != nil {
		return nil, err
	}

	// Gesamtwert des Lagers
	totalStockValue, err := s.articleRepo.CalculateTotalStockValue()
	if err != nil {
		return nil, err
	}

	// Transaktionen der letzten 30 Tage
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentTransactions, err := s.transactionRepo.CountSince(thirtyDaysAgo)
	if err != nil {
		return nil, err
	}

	// Statistiken zum Bestandsabbau/-aufbau
	stockMovementData, err := s.transactionRepo.GetStockMovementSummary()
	if err != nil {
		stockMovementData = map[string][]float64{}
	}

	return map[string]interface{}{
		"totalArticles":      totalArticles,
		"lowStockCount":      len(lowStockArticles),
		"totalStockValue":    totalStockValue,
		"recentTransactions": recentTransactions,
		"stockMovementData":  stockMovementData,
		"updateTimestamp":    time.Now(),
	}, nil
}
