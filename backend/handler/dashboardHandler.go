// backend/handler/dashboardHandler.go
package handler

import (
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DashboardHandler verwaltet alle Dashboard-bezogenen Anfragen
type DashboardHandler struct {
	articleRepo     *repository.ArticleRepository
	supplierRepo    *repository.SupplierRepository
	transactionRepo *repository.TransactionRepository
	activityRepo    *repository.ActivityRepository
}

// NewDashboardHandler erstellt einen neuen DashboardHandler
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		articleRepo:     repository.NewArticleRepository(),
		supplierRepo:    repository.NewSupplierRepository(),
		transactionRepo: repository.NewTransactionRepository(),
		activityRepo:    repository.NewActivityRepository(),
	}
}

// ShowDashboard zeigt das Dashboard an
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole := c.GetString("userRole")

	if userRole == string(model.RoleUser) {
		// Eingeschränktes Dashboard für normale Benutzer
		h.showLimitedDashboard(c, userModel)
		return
	}

	// Volles Dashboard für Admins/Manager
	// Statistiken abrufen
	stats, err := h.getWarehouseStats()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Dashboard-Daten: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Artikel unter Mindestbestand
	lowStockArticles, err := h.articleRepo.FindLowStock(10) // Die Top 10
	if err != nil {
		lowStockArticles = []*model.Article{} // Leere Liste im Fehlerfall
	}

	// Neueste Transaktionen
	recentTransactions, err := h.transactionRepo.FindRecent(5)
	if err != nil {
		recentTransactions = []*model.Transaction{} // Leere Liste im Fehlerfall
	}

	// Neueste Aktivitäten
	recentActivitiesData, err := h.activityRepo.FindRecent(10)
	if err != nil {
		recentActivitiesData = []*model.Activity{} // Leere Liste im Fehlerfall
	}

	// Aktivitäten in ein Template-freundlicheres Format konvertieren
	var recentActivities []gin.H
	for i, activity := range recentActivitiesData {
		recentActivities = append(recentActivities, gin.H{
			"IconBgClass": activity.GetIconClass(),
			"IconSVG":     activity.GetIconSVG(),
			"Message":     formatActivityMessage(activity),
			"Time":        activity.FormatTimeAgo(),
			"IsLast":      i == len(recentActivitiesData)-1,
		})
	}

	// Transaktionsdaten für Diagramme
	stockMovementData, err := h.transactionRepo.GetStockMovementSummary()
	if err != nil {
		stockMovementData = map[string][]float64{
			"stockIn":    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			"stockOut":   {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			"adjustment": {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	}

	// Und dann separate Labels definieren
	monthLabels := []string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}

	// Verteilung nach Kategorien
	categorySummary, err := h.articleRepo.GetCategorySummary()
	if err != nil {
		categorySummary = map[string][]interface{}{
			"labels": {"Keine Kategorie"},
			"counts": {0},
			"values": {0.0},
		}
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":              "Dashboard",
		"active":             "dashboard",
		"user":               userModel.FirstName + " " + userModel.LastName,
		"email":              userModel.Email,
		"year":               time.Now().Year(),
		"userRole":           userRole,
		"stats":              stats,
		"lowStockArticles":   lowStockArticles,
		"recentTransactions": recentTransactions,
		"recentActivities":   recentActivities,
		"stockMovementData":  stockMovementData,
		"monthLabels":        monthLabels,
		"categoryLabels":     categorySummary["labels"],
		"categoryCounts":     categorySummary["counts"],
		"categoryValues":     categorySummary["values"],
	})
}

// showLimitedDashboard zeigt ein eingeschränktes Dashboard für normale Benutzer
func (h *DashboardHandler) showLimitedDashboard(c *gin.Context, userModel *model.User) {
	// Neueste Aktivitäten dieses Benutzers
	recentActivitiesData, err := h.activityRepo.FindByUserID(userModel.ID.Hex(), 5)
	if err != nil {
		recentActivitiesData = []*model.Activity{} // Leere Liste im Fehlerfall
	}

	// Aktivitäten in ein einfacheres Format konvertieren
	var recentActivities []gin.H
	for _, activity := range recentActivitiesData {
		recentActivities = append(recentActivities, gin.H{
			"Time":    activity.Timestamp.Format("02.01.2006 15:04"),
			"Message": formatActivityMessage(activity),
		})
	}

	// Einfache Statistiken anzeigen
	articleCount, _ := h.articleRepo.Count()
	lowStockCount, _ := h.articleRepo.CountLowStock()

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":            "Dashboard",
		"active":           "dashboard",
		"user":             userModel.FirstName + " " + userModel.LastName,
		"email":            userModel.Email,
		"year":             time.Now().Year(),
		"userRole":         "user",
		"recentActivities": recentActivities,
		"articleCount":     articleCount,
		"lowStockCount":    lowStockCount,
	})
}

// getWarehouseStats sammelt verschiedene Lagerstatistiken
func (h *DashboardHandler) getWarehouseStats() (map[string]interface{}, error) {
	// Artikelzählung
	articleCount, err := h.articleRepo.Count()
	if err != nil {
		return nil, err
	}

	// Anzahl der Artikel unter Mindestbestand
	lowStockCount, err := h.articleRepo.CountLowStock()
	if err != nil {
		return nil, err
	}

	// Anzahl der Lieferanten
	supplierCount, err := h.supplierRepo.Count()
	if err != nil {
		return nil, err
	}

	// Wert des Gesamtbestands
	totalStockValue, err := h.articleRepo.CalculateTotalStockValue()
	if err != nil {
		return nil, err
	}

	// Anzahl der Kategorien
	categoryCount, err := h.articleRepo.CountCategories()
	if err != nil {
		return nil, err
	}

	// Bestandsmenge
	totalStock, err := h.articleRepo.CalculateTotalStock()
	if err != nil {
		return nil, err
	}

	// Transaktionen der letzten 30 Tage
	recentTransactionsCount, err := h.transactionRepo.CountSince(time.Now().AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"articleCount":            articleCount,
		"lowStockCount":           lowStockCount,
		"supplierCount":           supplierCount,
		"totalStockValue":         totalStockValue,
		"categoryCount":           categoryCount,
		"totalStock":              totalStock,
		"recentTransactionsCount": recentTransactionsCount,
		"lastUpdateTime":          time.Now().Format("02.01.2006 15:04"),
	}, nil
}

// formatActivityMessage formatiert die Aktivitätsnachricht für die Anzeige
func formatActivityMessage(activity *model.Activity) string {
	switch activity.Type {
	case model.ActivityTypeArticleAdded:
		return "<span class=\"font-medium text-gray-900\">" + activity.TargetName + "</span> wurde als neuer Artikel hinzugefügt"
	case model.ActivityTypeArticleUpdated:
		return "Artikel <span class=\"font-medium text-gray-900\">" + activity.TargetName + "</span> wurde aktualisiert"
	case model.ActivityTypeArticleDeleted:
		return "Artikel <span class=\"font-medium text-gray-900\">" + activity.TargetName + "</span> wurde gelöscht"
	case model.ActivityTypeStockAdjusted:
		return "Bestand für <span class=\"font-medium text-gray-900\">" + activity.TargetName + "</span> wurde angepasst"
	case model.ActivityTypeStockTaking:
		return "Inventur für <span class=\"font-medium text-gray-900\">" + activity.TargetName + "</span> wurde durchgeführt"
	case model.ActivityTypeUserLogin:
		return "<span class=\"font-medium text-gray-900\">" + activity.UserName + "</span> hat sich angemeldet"
	default:
		return activity.Description
	}
}
