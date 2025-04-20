// backend/model/supplier.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Supplier repräsentiert einen Lieferanten
type Supplier struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SupplierCode  string             `bson:"supplierCode" json:"supplierCode"`   // Interne Lieferantennummer
	Name          string             `bson:"name" json:"name"`                   // Firmenname
	ContactPerson string             `bson:"contactPerson" json:"contactPerson"` // Ansprechpartner
	Email         string             `bson:"email" json:"email"`                 // E-Mail-Adresse
	Phone         string             `bson:"phone" json:"phone"`                 // Telefonnummer
	Address       string             `bson:"address" json:"address"`             // Anschrift
	Website       string             `bson:"website,omitempty" json:"website,omitempty"`
	TaxID         string             `bson:"taxId,omitempty" json:"taxId,omitempty"` // Steuernummer
	PaymentTerms  string             `bson:"paymentTerms" json:"paymentTerms"`       // Zahlungsbedingungen
	Notes         string             `bson:"notes,omitempty" json:"notes,omitempty"` // Bemerkungen
	IsActive      bool               `bson:"isActive" json:"isActive"`               // Aktiv/Inaktiv
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
