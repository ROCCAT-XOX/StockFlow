// backend/model/article.go (erweiterte Version)
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Article repräsentiert einen Artikel im Lager
type Article struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ArticleNumber         string             `bson:"articleNumber" json:"articleNumber"`                 // Artikelnummer (eindeutig)
	ShortName             string             `bson:"shortName" json:"shortName"`                         // Kurztitel
	LongName              string             `bson:"longName" json:"longName"`                           // Detailbeschreibung
	EAN                   string             `bson:"ean" json:"ean"`                                     // Barcode
	Category              string             `bson:"category" json:"category"`                           // Warengruppe
	Unit                  string             `bson:"unit" json:"unit"`                                   // Lagereinheit (z.B. Stück, kg)
	StockCurrent          float64            `bson:"stockCurrent" json:"stockCurrent"`                   // Aktueller Lagerbestand
	StockReserved         float64            `bson:"stockReserved" json:"stockReserved"`                 // Reservierte Menge
	MinimumStock          float64            `bson:"minimumStock" json:"minimumStock"`                   // Mindestbestand/Bestellpunkt
	MaximumStock          float64            `bson:"maximumStock" json:"maximumStock"`                   // Maximalbestand (neu)
	ReorderQuantity       float64            `bson:"reorderQuantity" json:"reorderQuantity"`             // Bestellmenge (neu)
	PurchasePriceNet      float64            `bson:"purchasePriceNet" json:"purchasePriceNet"`           // Einkaufspreis (netto)
	SalesPriceGross       float64            `bson:"salesPriceGross" json:"salesPriceGross"`             // Verkaufspreis (brutto)
	SupplierID            primitive.ObjectID `bson:"supplierId,omitempty" json:"supplierId,omitempty"`   // Referenz zum Lieferanten (neu)
	SupplierNumber        string             `bson:"supplierNumber" json:"supplierNumber"`               // Lieferantennummer
	SupplierArticleNumber string             `bson:"supplierArticleNumber" json:"supplierArticleNumber"` // Artikelnummer beim Lieferanten (neu)
	DeliveryTimeInDays    int                `bson:"deliveryTimeInDays" json:"deliveryTimeInDays"`       // Lieferzeit in Tagen
	StorageLocation       string             `bson:"storageLocation" json:"storageLocation"`             // Lagerort
	Bin                   string             `bson:"bin" json:"bin"`                                     // Fach/Regal (neu)
	WeightKg              float64            `bson:"weightKg" json:"weightKg"`                           // Gewicht in kg
	Dimensions            string             `bson:"dimensions" json:"dimensions"`                       // Abmessungen (LxBxH) in cm
	SerialNumberRequired  bool               `bson:"serialNumberRequired" json:"serialNumberRequired"`   // Seriennummernpflicht
	HazardClass           string             `bson:"hazardClass" json:"hazardClass"`                     // Gefahrgutklasse
	Notes                 string             `bson:"notes" json:"notes"`                                 // Bemerkungen
	Images                []string           `bson:"images,omitempty" json:"images,omitempty"`           // Bilder (neu)
	IsActive              bool               `bson:"isActive" json:"isActive"`                           // Aktiv/Inaktiv (neu)
	LastStockTakeDate     time.Time          `bson:"lastStockTakeDate" json:"lastStockTakeDate"`         // Letztes Inventurdatum (neu)
	CreatedAt             time.Time          `bson:"createdAt" json:"createdAt"`                         // Erstellungsdatum
	UpdatedAt             time.Time          `bson:"updatedAt" json:"updatedAt"`                         // Aktualisierungsdatum
}

// GetStockStatus gibt den Bestandsstatus zurück (zu niedrig, optimal, zu hoch)
func (a *Article) GetStockStatus() string {
	if a.StockCurrent <= a.MinimumStock {
		return "low"
	} else if a.MaximumStock > 0 && a.StockCurrent >= a.MaximumStock {
		return "high"
	}
	return "ok"
}

// GetAvailableStock gibt den verfügbaren Bestand zurück (aktuell - reserviert)
func (a *Article) GetAvailableStock() float64 {
	return a.StockCurrent - a.StockReserved
}

// IsBelowMinimum prüft, ob der Bestand unter dem Mindestbestand ist
func (a *Article) IsBelowMinimum() bool {
	return a.StockCurrent <= a.MinimumStock
}

// GetStockValue gibt den aktuellen Warenwert zurück (Bestand * Einkaufspreis)
func (a *Article) GetStockValue() float64 {
	return a.StockCurrent * a.PurchasePriceNet
}
