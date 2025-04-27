// backend/handler/locationHandler.go
package handler

import (
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"

	"StockFlow/backend/model"
	"StockFlow/backend/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LocationHandler verwaltet alle Anfragen zu Lagerorten
type LocationHandler struct {
	locationRepo *repository.LocationRepository
}

// NewLocationHandler erstellt einen neuen LocationHandler
func NewLocationHandler() *LocationHandler {
	return &LocationHandler{
		locationRepo: repository.NewLocationRepository(),
	}
}

// ListLocations zeigt die Liste aller Lagerorte an
func (h *LocationHandler) ListLocations(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Alle Lagerorte abrufen
	locations, err := h.locationRepo.FindAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Lagerorte: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Lagerorte in eine hierarchische Struktur umwandeln
	locationMap := make(map[primitive.ObjectID]*model.Location)
	for _, loc := range locations {
		locationMap[loc.ID] = loc
	}

	// Hauptlager (ohne Parent) identifizieren
	var warehouses []*model.Location
	for _, loc := range locations {
		if loc.Type == model.LocationTypeWarehouse {
			warehouses = append(warehouses, loc)
		}
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "locations.html", gin.H{
		"title":      "Lagerorte",
		"active":     "locations",
		"user":       userModel.FirstName + " " + userModel.LastName,
		"email":      userModel.Email,
		"year":       time.Now().Year(),
		"locations":  locations,
		"warehouses": warehouses,
		"locMap":     locationMap,
		"userRole":   c.GetString("userRole"),
	})
}

// ShowAddLocationForm zeigt das Formular zum Hinzufügen eines Lagerorts an
func (h *LocationHandler) ShowAddLocationForm(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Alle Lagerorte für die Auswahl des übergeordneten Lagerorts laden
	locations, err := h.locationRepo.FindAll()
	if err != nil {
		locations = []*model.Location{} // Leere Liste im Fehlerfall
	}

	// Lagertyp aus der URL lesen (optional)
	locationType := c.Query("type")
	if locationType == "" {
		locationType = string(model.LocationTypeWarehouse)
	}

	// Parent-ID aus der URL lesen (optional)
	parentID := c.Query("parent")

	c.HTML(http.StatusOK, "location_add.html", gin.H{
		"title":     "Lagerort hinzufügen",
		"active":    "locations",
		"user":      userModel.FirstName + " " + userModel.LastName,
		"email":     userModel.Email,
		"year":      time.Now().Year(),
		"locations": locations,
		"parentID":  parentID,
		"locType":   locationType,
		"userRole":  c.GetString("userRole"),
	})
}

// AddLocation fügt einen neuen Lagerort hinzu
func (h *LocationHandler) AddLocation(c *gin.Context) {
	// Formulardaten abrufen
	name := c.PostForm("name")
	locationType := c.PostForm("type")
	description := c.PostForm("description")
	address := c.PostForm("address")
	parentID := c.PostForm("parentId")
	capacity := 0.0 // Optional, könnte aus dem Formular kommen

	// Parent ID als ObjectID konvertieren, falls vorhanden
	var parentObjID primitive.ObjectID
	if parentID != "" {
		var err error
		parentObjID, err = primitive.ObjectIDFromHex(parentID)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"title":   "Fehler",
				"message": "Ungültige übergeordnete ID",
				"year":    time.Now().Year(),
			})
			return
		}
	}

	// Neuen Lagerort erstellen
	location := &model.Location{
		Name:        name,
		Type:        model.LocationType(locationType),
		Description: description,
		Address:     address,
		ParentID:    parentObjID,
		IsActive:    true,
		Capacity:    capacity,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Lagerort in der Datenbank speichern
	err := h.locationRepo.Create(location)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Erstellen des Lagerorts: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleAdded, // Hier könnte ein spezieller Aktivitätstyp definiert werden
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		location.ID,
		"location",
		location.Name,
		"Neuer Lagerort hinzugefügt",
		0,
	)

	// Zurück zur Lagerortliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/locations?success=added")
}

// ShowEditLocationForm zeigt das Formular zum Bearbeiten eines Lagerorts an
func (h *LocationHandler) ShowEditLocationForm(c *gin.Context) {
	id := c.Param("id")

	// Lagerort anhand der ID abrufen
	location, err := h.locationRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Lagerort nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Alle Lagerorte für die Auswahl des übergeordneten Lagerorts laden
	locations, err := h.locationRepo.FindAll()
	if err != nil {
		locations = []*model.Location{} // Leere Liste im Fehlerfall
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	c.HTML(http.StatusOK, "location_edit.html", gin.H{
		"title":     "Lagerort bearbeiten",
		"active":    "locations",
		"user":      userModel.FirstName + " " + userModel.LastName,
		"email":     userModel.Email,
		"year":      time.Now().Year(),
		"location":  location,
		"locations": locations,
		"userRole":  c.GetString("userRole"),
	})
}

// UpdateLocation aktualisiert einen bestehenden Lagerort
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")

	// Lagerort anhand der ID abrufen
	location, err := h.locationRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Lagerort nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Formulardaten abrufen und Lagerort aktualisieren
	location.Name = c.PostForm("name")
	location.Type = model.LocationType(c.PostForm("type"))
	location.Description = c.PostForm("description")
	location.Address = c.PostForm("address")

	// Parent ID als ObjectID konvertieren, falls vorhanden
	parentID := c.PostForm("parentId")
	if parentID != "" {
		parentObjID, err := primitive.ObjectIDFromHex(parentID)
		if err == nil { // Nur setzen, wenn die Konvertierung erfolgreich war
			location.ParentID = parentObjID
		}
	} else {
		location.ParentID = primitive.NilObjectID
	}

	location.IsActive = c.PostForm("isActive") == "on"
	location.UpdatedAt = time.Now()

	// Lagerort in der Datenbank aktualisieren
	err = h.locationRepo.Update(location)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Lagerorts: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleUpdated, // Hier könnte ein spezieller Aktivitätstyp definiert werden
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		location.ID,
		"location",
		location.Name,
		"Lagerort aktualisiert",
		0,
	)

	// Zurück zur Lagerortliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/locations?success=updated")
}

// DeleteLocation löscht einen Lagerort
func (h *LocationHandler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")

	// Lagerort anhand der ID abrufen für das Aktivitätsprotokoll
	location, err := h.locationRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lagerort nicht gefunden"})
		return
	}

	// Lagerort löschen
	err = h.locationRepo.Delete(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dieser Lagerort hat untergeordnete Elemente und kann nicht gelöscht werden"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Lagerorts: " + err.Error()})
		}
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeArticleDeleted, // Hier könnte ein spezieller Aktivitätstyp definiert werden
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		location.ID,
		"location",
		location.Name,
		"Lagerort gelöscht",
		0,
	)

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{"message": "Lagerort erfolgreich gelöscht"})
}

// GetLocationChildren gibt alle Kindelemente eines Lagerorts zurück (für AJAX-Anfragen)
func (h *LocationHandler) GetLocationChildren(c *gin.Context) {
	parentID := c.Param("id")

	// Kindelemente abrufen
	children, err := h.locationRepo.FindByParentID(parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Unterkategorien: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, children)
}
