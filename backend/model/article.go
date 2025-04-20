// backend/model/article.go
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article repräsentiert einen Artikel im Lager
type Article struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ArticleNumber        string             `bson:"articleNumber" json:"articleNumber"`               // Artikelnummer (eindeutig)
	ShortName            string             `bson:"shortName" json:"shortName"`                       // Kurztitel
	LongName             string             `bson:"longName" json:"longName"`                         // Detailbeschreibung
	EAN                  string             `bson:"ean" json:"ean"`                                   // Barcode
	Category             string             `bson:"category" json:"category"`                         // Warengruppe
	Unit                 string             `bson:"unit" json:"unit"`                                 // Lagereinheit (z.B. Stück, kg)
	StockCurrent         float64            `bson:"stockCurrent" json:"stockCurrent"`                 // Aktueller Lagerbestand
	StockReserved        float64            `bson:"stockReserved" json:"stockReserved"`               // Reservierte Menge
	MinimumStock         float64            `bson:"minimumStock" json:"minimumStock"`                 // Mindestbestand/Bestellpunkt
	PurchasePriceNet     float64            `bson:"purchasePriceNet" json:"purchasePriceNet"`         // Einkaufspreis (netto)
	SalesPriceGross      float64            `bson:"salesPriceGross" json:"salesPriceGross"`           // Verkaufspreis (brutto)
	SupplierNumber       string             `bson:"supplierNumber" json:"supplierNumber"`             // Lieferantennummer
	DeliveryTimeInDays   int                `bson:"deliveryTimeInDays" json:"deliveryTimeInDays"`     // Lieferzeit in Tagen
	StorageLocation      string             `bson:"storageLocation" json:"storageLocation"`           // Lagerort
	WeightKg             float64            `bson:"weightKg" json:"weightKg"`                         // Gewicht in kg
	Dimensions           string             `bson:"dimensions" json:"dimensions"`                     // Abmessungen (LxBxH) in cm
	SerialNumberRequired bool               `bson:"serialNumberRequired" json:"serialNumberRequired"` // Seriennummernpflicht
	HazardClass          string             `bson:"hazardClass" json:"hazardClass"`                   // Gefahrgutklasse
	Notes                string             `bson:"notes" json:"notes"`                               // Bemerkungen
	CreatedAt            time.Time          `bson:"createdAt" json:"createdAt"`                       // Erstellungsdatum
	UpdatedAt            time.Time          `bson:"updatedAt" json:"updatedAt"`                       // Aktualisierungsdatum
}
