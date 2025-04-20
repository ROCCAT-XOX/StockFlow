// backend/model/transaction.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// TransactionType repräsentiert den Typ einer Transaktion
type TransactionType string

const (
	TransactionTypeStockIn   TransactionType = "stock_in"  // Wareneingang
	TransactionTypeStockOut  TransactionType = "stock_out" // Warenausgang
	TransactionTypeAdjust    TransactionType = "adjust"    // Bestandskorrektur
	TransactionTypeInventory TransactionType = "inventory" // Inventurzählung
)

// Transaction repräsentiert eine Lager-Transaktion (Ein-/Ausgang/Korrektur)
type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        TransactionType    `bson:"type" json:"type"`
	ArticleID   primitive.ObjectID `bson:"articleId" json:"articleId"`
	ArticleName string             `bson:"articleName" json:"articleName"`
	Quantity    float64            `bson:"quantity" json:"quantity"`                       // Menge (positiv oder negativ)
	OldStock    float64            `bson:"oldStock" json:"oldStock"`                       // Bestand vor der Transaktion
	NewStock    float64            `bson:"newStock" json:"newStock"`                       // Bestand nach der Transaktion
	UnitPrice   float64            `bson:"unitPrice,omitempty" json:"unitPrice,omitempty"` // Stückpreis für Bewertung
	Reason      string             `bson:"reason,omitempty" json:"reason,omitempty"`       // Grund der Transaktion
	Reference   string             `bson:"reference,omitempty" json:"reference,omitempty"` // Referenz (z.B. Lieferschein, Bestellung)
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`                           // Benutzer, der die Transaktion durchgeführt hat
	UserName    string             `bson:"userName" json:"userName"`                       // Name des Benutzers für die Anzeige
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`                     // Zeitpunkt der Transaktion
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
}

// GetStatusClass gibt eine CSS-Klasse basierend auf dem Transaktionstyp zurück
func (t *Transaction) GetStatusClass() string {
	switch t.Type {
	case TransactionTypeStockIn:
		return "bg-green-100 text-green-800"
	case TransactionTypeStockOut:
		return "bg-red-100 text-red-800"
	case TransactionTypeAdjust:
		return "bg-yellow-100 text-yellow-800"
	case TransactionTypeInventory:
		return "bg-blue-100 text-blue-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
}

// GetDisplayType gibt einen benutzerfreundlichen Namen für den Transaktionstyp zurück
func (t *Transaction) GetDisplayType() string {
	switch t.Type {
	case TransactionTypeStockIn:
		return "Wareneingang"
	case TransactionTypeStockOut:
		return "Warenausgang"
	case TransactionTypeAdjust:
		return "Bestandskorrektur"
	case TransactionTypeInventory:
		return "Inventur"
	default:
		return string(t.Type)
	}
}
