// backend/model/location.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// LocationType repräsentiert den Typ eines Lagerorts
type LocationType string

const (
	LocationTypeWarehouse LocationType = "warehouse" // Lager/Hauptstandort
	LocationTypeArea      LocationType = "area"      // Bereich/Regal
	LocationTypeShelf     LocationType = "shelf"     // Fach
)

// Location repräsentiert einen Lagerort im System
type Location struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`                   // Name des Lagerorts
	Type        LocationType       `bson:"type" json:"type"`                   // Typ des Lagerorts
	Description string             `bson:"description" json:"description"`     // Optionale Beschreibung
	Address     string             `bson:"address,omitempty" json:"address"`   // Adresse (nur für Hauptlager)
	ParentID    primitive.ObjectID `bson:"parentId,omitempty" json:"parentId"` // Übergeordneter Lagerort (leer bei Hauptlagern)
	IsActive    bool               `bson:"isActive" json:"isActive"`           // Status des Lagerorts
	Capacity    float64            `bson:"capacity" json:"capacity"`           // Optionale Kapazitätsangabe
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// GetFullPath gibt den vollständigen Pfad des Lagerorts zurück
func (l *Location) GetFullPath(locations map[primitive.ObjectID]*Location) string {
	if l.ParentID.IsZero() {
		return l.Name
	}

	parent, exists := locations[l.ParentID]
	if !exists {
		return l.Name
	}

	return parent.GetFullPath(locations) + " > " + l.Name
}
