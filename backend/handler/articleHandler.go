// backend/handler/articleHandler.go
package handler

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
	"time"

	"StockFlow/backend/model"
	"StockFlow/backend/repository"

	"github.com/gin-gonic/gin"
)

// ArticleHandler verwaltet alle Anfragen zu Artikeln
type ArticleHandler struct {
	articleRepo  *repository.ArticleRepository
	supplierRepo *repository.SupplierRepository
}

// NewArticleHandler erstellt einen neuen ArticleHandler
func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{
		articleRepo:  repository.NewArticleRepository(),
		supplierRepo: repository.NewSupplierRepository(),
	}
}

// ListArticles zeigt die Liste aller Artikel an
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Alle Artikel abrufen
	articles, err := h.articleRepo.FindAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Artikel: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "articles.html", gin.H{
		"title":         "Artikel",
		"active":        "articles",
		"user":          userModel.FirstName + " " + userModel.LastName,
		"email":         userModel.Email,
		"year":          time.Now().Year(),
		"articles":      articles,
		"totalArticles": len(articles),
		"userRole":      c.GetString("userRole"),
	})
}

// ShowAddArticleForm zeigt das Formular zum Hinzufügen eines Artikels an
func (h *ArticleHandler) ShowAddArticleForm(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Nächste Artikelnummer generieren (für Formularvorausfüllung)
	nextArticleNumber, _ := h.articleRepo.GetNextArticleNumber()

	// Lieferanten für Dropdown abrufen
	suppliers, err := h.supplierRepo.FindAll()
	if err != nil {
		suppliers = []*model.Supplier{} // Leere Liste im Fehlerfall
	}

	c.HTML(http.StatusOK, "article_add.html", gin.H{
		"title":             "Artikel hinzufügen",
		"active":            "articles",
		"user":              userModel.FirstName + " " + userModel.LastName,
		"email":             userModel.Email,
		"year":              time.Now().Year(),
		"userRole":          c.GetString("userRole"),
		"suppliers":         suppliers,
		"nextArticleNumber": nextArticleNumber,
	})
}

// AddArticle fügt einen neuen Artikel hinzu
func (h *ArticleHandler) AddArticle(c *gin.Context) {
	// Formulardaten abrufen
	articleNumber := c.PostForm("articleNumber")
	shortName := c.PostForm("shortName")
	longName := c.PostForm("longName")
	ean := c.PostForm("ean")
	category := c.PostForm("category")
	unit := c.PostForm("unit")
	supplierIDStr := c.PostForm("supplierId")
	supplierArticleNumber := c.PostForm("supplierArticleNumber")
	maximumStockStr := c.PostForm("maximumStock")
	reorderQuantityStr := c.PostForm("reorderQuantity")
	bin := c.PostForm("bin")
	isActive := c.PostForm("isActive") == "on"

	// Zusätzliche Werte parsen
	var supplierID primitive.ObjectID
	if supplierIDStr != "" {
		supplierID, _ = primitive.ObjectIDFromHex(supplierIDStr)
	}

	maximumStock, _ := strconv.ParseFloat(maximumStockStr, 64)
	reorderQuantity, _ := strconv.ParseFloat(reorderQuantityStr, 64)

	// Numerische Werte parsen
	stockCurrent, _ := strconv.ParseFloat(c.PostForm("stockCurrent"), 64)
	stockReserved, _ := strconv.ParseFloat(c.PostForm("stockReserved"), 64)
	minimumStock, _ := strconv.ParseFloat(c.PostForm("minimumStock"), 64)
	purchasePriceNet, _ := strconv.ParseFloat(c.PostForm("purchasePriceNet"), 64)
	salesPriceGross, _ := strconv.ParseFloat(c.PostForm("salesPriceGross"), 64)
	deliveryTimeInDays, _ := strconv.Atoi(c.PostForm("deliveryTimeInDays"))
	weightKg, _ := strconv.ParseFloat(c.PostForm("weightKg"), 64)

	// Boolean-Wert parsen
	serialNumberRequired := c.PostForm("serialNumberRequired") == "on"

	// Neuen Artikel erstellen
	article := &model.Article{
		ArticleNumber:         articleNumber, // Wird automatisch überschrieben, wenn bereits vorhanden oder leer
		ShortName:             shortName,
		LongName:              longName,
		EAN:                   ean,
		Category:              category,
		Unit:                  unit,
		StockCurrent:          stockCurrent,
		StockReserved:         stockReserved,
		MinimumStock:          minimumStock,
		PurchasePriceNet:      purchasePriceNet,
		SalesPriceGross:       salesPriceGross,
		SupplierNumber:        c.PostForm("supplierNumber"),
		DeliveryTimeInDays:    deliveryTimeInDays,
		StorageLocation:       c.PostForm("storageLocation"),
		WeightKg:              weightKg,
		Dimensions:            c.PostForm("dimensions"),
		SerialNumberRequired:  serialNumberRequired,
		HazardClass:           c.PostForm("hazardClass"),
		Notes:                 c.PostForm("notes"),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		SupplierID:            supplierID,
		SupplierArticleNumber: supplierArticleNumber,
		MaximumStock:          maximumStock,
		ReorderQuantity:       reorderQuantity,
		Bin:                   bin,
		IsActive:              isActive,
		LastStockTakeDate:     time.Time{},
	}

	// Artikel in der Datenbank speichern
	err := h.articleRepo.Create(article)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Erstellen des Artikels: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleAdded,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		article.ID,
		"article",
		article.ShortName,
		"Neuer Artikel hinzugefügt",
		0, // Quantity-Parameter hinzufügen
	)

	// Zurück zur Artikelliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/articles?success=added")
}

// Neue Funktion zur Unterstützung der Bestandsübersicht
func (h *ArticleHandler) ShowStockOverview(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Filter-Parameter aus der Anfrage extrahieren
	categoryFilter := c.DefaultQuery("category", "")
	stockStatus := c.DefaultQuery("status", "")

	// Artikel abrufen (gefiltert, falls notwendig)
	var articles []*model.Article
	var err error

	if categoryFilter != "" && stockStatus != "" {
		articles, err = h.articleRepo.FindByCategoryAndStockStatus(categoryFilter, stockStatus)
	} else if categoryFilter != "" {
		articles, err = h.articleRepo.FindByCategory(categoryFilter)
	} else if stockStatus != "" {
		articles, err = h.articleRepo.FindByStockStatus(stockStatus)
	} else {
		articles, err = h.articleRepo.FindAll()
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Artikel: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Alle Kategorien für Filter
	categories, _ := h.articleRepo.GetAllCategories()

	// Statistiken berechnen
	var totalStock float64
	var totalValue float64
	var lowStockCount int

	for _, article := range articles {
		totalStock += article.StockCurrent
		totalValue += article.StockCurrent * article.PurchasePriceNet

		if article.IsBelowMinimum() {
			lowStockCount++
		}
	}

	c.HTML(http.StatusOK, "stock_overview.html", gin.H{
		"title":          "Bestandsübersicht",
		"active":         "stock",
		"user":           userModel.FirstName + " " + userModel.LastName,
		"email":          userModel.Email,
		"year":           time.Now().Year(),
		"articles":       articles,
		"categories":     categories,
		"categoryFilter": categoryFilter,
		"stockStatus":    stockStatus,
		"totalArticles":  len(articles),
		"totalStock":     totalStock,
		"totalValue":     totalValue,
		"lowStockCount":  lowStockCount,
		"userRole":       c.GetString("userRole"),
	})
}

// GetArticleDetails zeigt die Details eines Artikels an
func (h *ArticleHandler) GetArticleDetails(c *gin.Context) {
	id := c.Param("id")

	// Artikel anhand der ID abrufen
	article, err := h.articleRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Artikel nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "article_detail.html", gin.H{
		"title":    article.ShortName,
		"active":   "articles",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"article":  article,
		"userRole": c.GetString("userRole"),
	})
}

// ShowEditArticleForm zeigt das Formular zum Bearbeiten eines Artikels an
func (h *ArticleHandler) ShowEditArticleForm(c *gin.Context) {
	id := c.Param("id")

	// Artikel anhand der ID abrufen
	article, err := h.articleRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Artikel nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	c.HTML(http.StatusOK, "article_edit.html", gin.H{
		"title":    "Artikel bearbeiten",
		"active":   "articles",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"article":  article,
		"userRole": c.GetString("userRole"),
	})
}

// UpdateArticle aktualisiert einen bestehenden Artikel
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	// Artikel anhand der ID abrufen
	article, err := h.articleRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Artikel nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Formulardaten abrufen und Artikel aktualisieren
	article.ArticleNumber = c.PostForm("articleNumber")
	article.ShortName = c.PostForm("shortName")
	article.LongName = c.PostForm("longName")
	article.EAN = c.PostForm("ean")
	article.Category = c.PostForm("category")
	article.Unit = c.PostForm("unit")

	// Numerische Werte parsen
	article.StockCurrent, _ = strconv.ParseFloat(c.PostForm("stockCurrent"), 64)
	article.StockReserved, _ = strconv.ParseFloat(c.PostForm("stockReserved"), 64)
	article.MinimumStock, _ = strconv.ParseFloat(c.PostForm("minimumStock"), 64)
	article.PurchasePriceNet, _ = strconv.ParseFloat(c.PostForm("purchasePriceNet"), 64)
	article.SalesPriceGross, _ = strconv.ParseFloat(c.PostForm("salesPriceGross"), 64)
	article.SupplierNumber = c.PostForm("supplierNumber")
	article.DeliveryTimeInDays, _ = strconv.Atoi(c.PostForm("deliveryTimeInDays"))
	article.StorageLocation = c.PostForm("storageLocation")
	article.WeightKg, _ = strconv.ParseFloat(c.PostForm("weightKg"), 64)
	article.Dimensions = c.PostForm("dimensions")
	article.SerialNumberRequired = c.PostForm("serialNumberRequired") == "on"
	article.HazardClass = c.PostForm("hazardClass")
	article.Notes = c.PostForm("notes")

	// Artikel in der Datenbank aktualisieren
	err = h.articleRepo.Update(article)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Artikels: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleUpdated,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		article.ID,
		"article",
		article.ShortName,
		"Artikel aktualisiert",
		0, // Quantity-Parameter hinzufügen
	)

	// Zurück zur Artikelliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/articles?success=updated")
}

// DeleteArticle löscht einen Artikel
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	// Artikel anhand der ID abrufen für das Aktivitätsprotokoll
	article, err := h.articleRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artikel nicht gefunden"})
		return
	}

	// Artikel löschen
	err = h.articleRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Artikels: " + err.Error()})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleDeleted,
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		article.ID,
		"article",
		article.ShortName,
		"Artikel gelöscht",
		0, // Quantity-Parameter hinzufügen
	)

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{"message": "Artikel erfolgreich gelöscht"})
}
