// backend/handler/transactionHandler.go
package handler

import (
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionHandler verwaltet alle Anfragen zu Lagertransaktionen
type TransactionHandler struct {
	transactionRepo *repository.TransactionRepository
	articleRepo     *repository.ArticleRepository
}

// NewTransactionHandler erstellt einen neuen TransactionHandler
func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: repository.NewTransactionRepository(),
		articleRepo:     repository.NewArticleRepository(),
	}
}

// ListTransactions zeigt die Liste aller Transaktionen an
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Filter-Parameter aus der Anfrage extrahieren
	articleID := c.DefaultQuery("articleId", "")
	transactionType := c.DefaultQuery("type", "")

	// Paginierung
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage := 20 // Anzahl der Einträge pro Seite

	// Transaktionen aus der Datenbank abrufen
	transactions, total, err := h.transactionRepo.FindWithFilters(articleID, transactionType, page, perPage)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Transaktionen: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Artikel abrufen, falls ein Filter gesetzt ist
	var filterArticle *model.Article
	if articleID != "" {
		filterArticle, _ = h.articleRepo.FindByID(articleID)
	}

	// Gesamtseitenzahl berechnen
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "transactions.html", gin.H{
		"title":         "Lagerbewegungen",
		"active":        "transactions",
		"user":          userModel.FirstName + " " + userModel.LastName,
		"email":         userModel.Email,
		"year":          time.Now().Year(),
		"transactions":  transactions,
		"filterArticle": filterArticle,
		"filterType":    transactionType,
		"currentPage":   page,
		"totalPages":    totalPages,
		"total":         total,
		"userRole":      c.GetString("userRole"),
	})
}

// ShowAddTransactionForm zeigt das Formular zum Hinzufügen einer Transaktion an
func (h *TransactionHandler) ShowAddTransactionForm(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Artikel-ID aus URL-Parameter extrahieren
	articleID := c.Query("articleId")

	// Transaktionstyp aus URL-Parameter extrahieren (Standard: Wareneingang)
	transactionType := c.DefaultQuery("type", string(model.TransactionTypeStockIn))

	// Artikel abrufen, falls eine ID angegeben wurde
	var article *model.Article
	if articleID != "" {
		var err error
		article, err = h.articleRepo.FindByID(articleID)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"title":   "Fehler",
				"message": "Artikel nicht gefunden",
				"year":    time.Now().Year(),
			})
			return
		}
	}

	// Alle Artikel für das Dropdown abrufen
	articles, err := h.articleRepo.FindAll()
	if err != nil {
		articles = []*model.Article{} // Leere Liste im Fehlerfall
	}

	c.HTML(http.StatusOK, "transaction_add.html", gin.H{
		"title":           "Lagerbewegung erfassen",
		"active":          "transactions",
		"user":            userModel.FirstName + " " + userModel.LastName,
		"email":           userModel.Email,
		"year":            time.Now().Year(),
		"articles":        articles,
		"selectedArticle": article,
		"type":            transactionType,
		"userRole":        c.GetString("userRole"),
	})
}

// AddTransaction fügt eine neue Transaktion hinzu und aktualisiert den Lagerbestand
func (h *TransactionHandler) AddTransaction(c *gin.Context) {
	// Daten aus dem Formular extrahieren
	articleIDStr := c.PostForm("articleId")
	transactionType := c.PostForm("type")
	quantityStr := c.PostForm("quantity")
	reason := c.PostForm("reason")
	reference := c.PostForm("reference")
	notes := c.PostForm("notes")
	unitPriceStr := c.PostForm("unitPrice")

	// ArticleID in ObjectID umwandeln
	articleID, err := primitive.ObjectIDFromHex(articleIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Ungültige Artikel-ID",
			"year":    time.Now().Year(),
		})
		return
	}

	// Artikel aus der Datenbank abrufen
	article, err := h.articleRepo.FindByID(articleIDStr)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Artikel nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Menge als Float parsen
	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil || quantity <= 0 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Ungültige Menge. Bitte geben Sie eine positive Zahl ein.",
			"year":    time.Now().Year(),
		})
		return
	}

	// Stückpreis als Float parsen (falls vorhanden)
	var unitPrice float64
	if unitPriceStr != "" {
		unitPrice, _ = strconv.ParseFloat(unitPriceStr, 64)
	} else {
		// Standardmäßig den Einkaufspreis des Artikels verwenden
		unitPrice = article.PurchasePriceNet
	}

	// Alten Bestand speichern
	oldStock := article.StockCurrent

	// Neuen Bestand berechnen basierend auf dem Transaktionstyp
	var newStock float64
	switch transactionType {
	case string(model.TransactionTypeStockIn):
		newStock = oldStock + quantity
	case string(model.TransactionTypeStockOut):
		newStock = oldStock - quantity
		// Prüfen, ob genügend Bestand vorhanden ist
		if newStock < 0 {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"title":   "Fehler",
				"message": "Nicht genügend Bestand vorhanden.",
				"year":    time.Now().Year(),
			})
			return
		}
	case string(model.TransactionTypeAdjust):
		// Bei einer Anpassung ist die eingegebene Menge bereits der neue Bestand
		newStock = quantity
		// Anpassen der Menge für die Transaktion (Differenz zum alten Bestand)
		quantity = newStock - oldStock
	case string(model.TransactionTypeInventory):
		// Bei einer Inventur ist die eingegebene Menge der gezählte Bestand
		newStock = quantity
		// Anpassen der Menge für die Transaktion (Differenz zum alten Bestand)
		quantity = newStock - oldStock
	default:
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Ungültiger Transaktionstyp",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Neue Transaktion erstellen
	transaction := &model.Transaction{
		ID:          primitive.NewObjectID(),
		Type:        model.TransactionType(transactionType),
		ArticleID:   articleID,
		ArticleName: article.ShortName,
		Quantity:    quantity,
		OldStock:    oldStock,
		NewStock:    newStock,
		UnitPrice:   unitPrice,
		Reason:      reason,
		Reference:   reference,
		UserID:      userModel.ID,
		UserName:    fmt.Sprintf("%s %s", userModel.FirstName, userModel.LastName),
		Timestamp:   time.Now(),
		Notes:       notes,
	}

	// Transaktion speichern
	err = h.transactionRepo.Create(transaction)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Speichern der Transaktion: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Artikelbestand aktualisieren
	article.StockCurrent = newStock
	if transactionType == string(model.TransactionTypeInventory) {
		article.LastStockTakeDate = time.Now()
	}

	err = h.articleRepo.Update(article)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Artikelbestands: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	activityRepo := repository.NewActivityRepository()
	activityType := model.ActivityTypeStockAdjusted
	if transactionType == string(model.TransactionTypeInventory) {
		activityType = model.ActivityTypeStockTaking
	}

	_, _ = activityRepo.LogActivity(
		activityType,
		userModel.ID,
		fmt.Sprintf("%s %s", userModel.FirstName, userModel.LastName),
		articleID,
		"article",
		article.ShortName,
		fmt.Sprintf("%s: %g %s (neu: %g)", transaction.GetDisplayType(), quantity, article.Unit, newStock),
		quantity,
	)

	// Weiterleitungsziel: zurück zur Artikel-Detailseite oder zur Transaktionsliste
	redirectURL := fmt.Sprintf("/articles/view/%s?success=transaction", articleIDStr)
	if c.PostForm("returnToList") == "true" {
		redirectURL = "/transactions?success=added"
	}

	// Weiterleitung mit Erfolgsmeldung
	c.Redirect(http.StatusFound, redirectURL)
}

// GetTransactionDetails zeigt die Details einer Transaktion an
func (h *TransactionHandler) GetTransactionDetails(c *gin.Context) {
	id := c.Param("id")

	// Transaktion anhand der ID abrufen
	transaction, err := h.transactionRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Transaktion nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "transaction_detail.html", gin.H{
		"title":       "Transaktionsdetails",
		"active":      "transactions",
		"user":        userModel.FirstName + " " + userModel.LastName,
		"email":       userModel.Email,
		"year":        time.Now().Year(),
		"transaction": transaction,
		"userRole":    c.GetString("userRole"),
	})
}
