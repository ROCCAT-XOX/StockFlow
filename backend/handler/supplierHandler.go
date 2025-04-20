// backend/handler/supplierHandler.go
package handler

import (
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SupplierHandler verwaltet alle Anfragen zu Lieferanten
type SupplierHandler struct {
	supplierRepo *repository.SupplierRepository
	articleRepo  *repository.ArticleRepository
}

// NewSupplierHandler erstellt einen neuen SupplierHandler
func NewSupplierHandler() *SupplierHandler {
	return &SupplierHandler{
		supplierRepo: repository.NewSupplierRepository(),
		articleRepo:  repository.NewArticleRepository(),
	}
}

// ListSuppliers zeigt die Liste aller Lieferanten an
func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Alle Lieferanten abrufen
	suppliers, err := h.supplierRepo.FindAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Lieferanten: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "suppliers.html", gin.H{
		"title":          "Lieferanten",
		"active":         "suppliers",
		"user":           userModel.FirstName + " " + userModel.LastName,
		"email":          userModel.Email,
		"year":           time.Now().Year(),
		"suppliers":      suppliers,
		"totalSuppliers": len(suppliers),
		"userRole":       c.GetString("userRole"),
	})
}

// ShowAddSupplierForm zeigt das Formular zum Hinzufügen eines Lieferanten an
func (h *SupplierHandler) ShowAddSupplierForm(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	c.HTML(http.StatusOK, "supplier_add.html", gin.H{
		"title":    "Lieferanten hinzufügen",
		"active":   "suppliers",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"userRole": c.GetString("userRole"),
	})
}

// AddSupplier fügt einen neuen Lieferanten hinzu
func (h *SupplierHandler) AddSupplier(c *gin.Context) {
	// Daten aus dem Formular extrahieren
	supplierCode := c.PostForm("supplierCode")
	name := c.PostForm("name")
	contactPerson := c.PostForm("contactPerson")
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	address := c.PostForm("address")
	website := c.PostForm("website")
	taxID := c.PostForm("taxID")
	paymentTerms := c.PostForm("paymentTerms")
	notes := c.PostForm("notes")

	// Erstellen eines neuen Lieferanten
	supplier := &model.Supplier{
		SupplierCode:  supplierCode,
		Name:          name,
		ContactPerson: contactPerson,
		Email:         email,
		Phone:         phone,
		Address:       address,
		Website:       website,
		TaxID:         taxID,
		PaymentTerms:  paymentTerms,
		Notes:         notes,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Lieferanten in der Datenbank speichern
	err := h.supplierRepo.Create(supplier)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Erstellen des Lieferanten: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeSupplierAdded, // Neuer Aktivitätstyp, müsste im Modell definiert werden
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		supplier.ID,
		"supplier",
		supplier.Name,
		"Neuer Lieferant hinzugefügt",
		0,
	)

	// Weiterleitung zur Lieferantenliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/suppliers?success=added")
}

// GetSupplierDetails zeigt die Details eines Lieferanten an
func (h *SupplierHandler) GetSupplierDetails(c *gin.Context) {
	id := c.Param("id")

	// Lieferanten anhand der ID abrufen
	supplier, err := h.supplierRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Lieferant nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Alle Artikel dieses Lieferanten finden
	articles, err := h.articleRepo.FindBySupplierID(supplier.ID.Hex())
	if err != nil {
		articles = []*model.Article{} // Leere Liste im Fehlerfall
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "supplier_detail.html", gin.H{
		"title":    supplier.Name,
		"active":   "suppliers",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"supplier": supplier,
		"articles": articles,
		"userRole": c.GetString("userRole"),
	})
}

// ShowEditSupplierForm zeigt das Formular zum Bearbeiten eines Lieferanten an
func (h *SupplierHandler) ShowEditSupplierForm(c *gin.Context) {
	id := c.Param("id")

	// Lieferanten anhand der ID abrufen
	supplier, err := h.supplierRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Lieferant nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	c.HTML(http.StatusOK, "supplier_edit.html", gin.H{
		"title":    "Lieferant bearbeiten",
		"active":   "suppliers",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"supplier": supplier,
		"userRole": c.GetString("userRole"),
	})
}

// UpdateSupplier aktualisiert einen bestehenden Lieferanten
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	id := c.Param("id")

	// Lieferanten anhand der ID abrufen
	supplier, err := h.supplierRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Lieferant nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Daten aus dem Formular extrahieren und Lieferanten aktualisieren
	supplier.SupplierCode = c.PostForm("supplierCode")
	supplier.Name = c.PostForm("name")
	supplier.ContactPerson = c.PostForm("contactPerson")
	supplier.Email = c.PostForm("email")
	supplier.Phone = c.PostForm("phone")
	supplier.Address = c.PostForm("address")
	supplier.Website = c.PostForm("website")
	supplier.TaxID = c.PostForm("taxID")
	supplier.PaymentTerms = c.PostForm("paymentTerms")
	supplier.Notes = c.PostForm("notes")
	supplier.IsActive = c.PostForm("isActive") == "on"
	supplier.UpdatedAt = time.Now()

	// Lieferanten in der Datenbank aktualisieren
	err = h.supplierRepo.Update(supplier)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Lieferanten: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeSupplierUpdated, // Neuer Aktivitätstyp, müsste im Modell definiert werden
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		supplier.ID,
		"supplier",
		supplier.Name,
		"Lieferant aktualisiert",
		0,
	)

	// Weiterleitung zur Lieferantenliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/suppliers?success=updated")
}

// DeleteSupplier löscht einen Lieferanten
func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	id := c.Param("id")

	// Prüfen, ob der Lieferant mit Artikeln verknüpft ist
	articles, err := h.articleRepo.FindBySupplierID(id)
	if err == nil && len(articles) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Dieser Lieferant ist mit %d Artikeln verknüpft und kann nicht gelöscht werden", len(articles)),
		})
		return
	}

	// Lieferanten anhand der ID abrufen (für das Aktivitätsprotokoll)
	supplier, err := h.supplierRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lieferant nicht gefunden"})
		return
	}

	// Lieferanten löschen
	err = h.supplierRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Lieferanten: " + err.Error()})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeSupplierDeleted, // Neuer Aktivitätstyp, müsste im Modell definiert werden
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		supplier.ID,
		"supplier",
		supplier.Name,
		"Lieferant gelöscht",
		0,
	)

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{"message": "Lieferant erfolgreich gelöscht"})
}
