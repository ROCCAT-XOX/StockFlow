// backend/repository/initRepository.go
package repository

import (
	"context"
	"log"
	"time"

	"StockFlow/backend/db"
	"StockFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
)

// InitRepository ist für die Initialisierung der Datenbank zuständig
type InitRepository struct {
	articleRepo *ArticleRepository
	userRepo    *UserRepository
}

// NewInitRepository erstellt ein neues InitRepository
func NewInitRepository() *InitRepository {
	return &InitRepository{
		articleRepo: NewArticleRepository(),
		userRepo:    NewUserRepository(),
	}
}

// InitializeDatabase initialisiert die Datenbank mit Beispieldaten, falls nötig
func (r *InitRepository) InitializeDatabase() error {
	// Zuerst Admin-Benutzer erstellen, falls keiner existiert
	err := r.userRepo.CreateAdminUserIfNotExists()
	if err != nil {
		log.Printf("Warnung: Admin-Benutzer konnte nicht erstellt werden: %v", err)
	} else {
		log.Println("Admin-Benutzer wurde überprüft/erstellt")
	}

	// Prüfen, ob bereits Artikel vorhanden sind
	count, err := r.countArticles()
	if err != nil {
		return err
	}

	// Wenn keine Artikel vorhanden sind, Beispielartikel erstellen
	if count == 0 {
		err = r.createSampleArticles()
		if err != nil {
			return err
		}
		log.Println("Beispielartikel wurden erfolgreich erstellt")
	}

	return nil
}

// countArticles zählt die Anzahl der Artikel in der Datenbank
func (r *InitRepository) countArticles() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.GetCollection("articles")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// createSampleArticles erstellt einige Beispielartikel in der Datenbank
func (r *InitRepository) createSampleArticles() error {
	sampleArticles := []model.Article{
		{
			ArticleNumber:        "A1001",
			ShortName:            "Hammer, 300g",
			LongName:             "Schlosshammer mit Hickorystiel, 300g",
			EAN:                  "4012345678901",
			Category:             "Werkzeug",
			Unit:                 "Stück",
			StockCurrent:         15,
			StockReserved:        2,
			MinimumStock:         5,
			PurchasePriceNet:     8.90,
			SalesPriceGross:      14.99,
			SupplierNumber:       "SUPPL-001",
			DeliveryTimeInDays:   3,
			StorageLocation:      "Regal A1",
			WeightKg:             0.35,
			Dimensions:           "30x12x4",
			SerialNumberRequired: false,
			HazardClass:          "",
			Notes:                "Bestseller, regelmäßig nachbestellen",
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ArticleNumber:        "A2002",
			ShortName:            "Schraubendreher-Set",
			LongName:             "Präzisions-Schraubendreher-Set, 6-teilig",
			EAN:                  "4012345678902",
			Category:             "Werkzeug",
			Unit:                 "Set",
			StockCurrent:         8,
			StockReserved:        0,
			MinimumStock:         3,
			PurchasePriceNet:     12.50,
			SalesPriceGross:      19.95,
			SupplierNumber:       "SUPPL-002",
			DeliveryTimeInDays:   5,
			StorageLocation:      "Regal A2",
			WeightKg:             0.25,
			Dimensions:           "20x10x3",
			SerialNumberRequired: false,
			HazardClass:          "",
			Notes:                "",
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ArticleNumber:        "B3003",
			ShortName:            "Akku-Bohrschrauber",
			LongName:             "Profi Akku-Bohrschrauber 18V, inkl. 2 Akkus und Ladegerät",
			EAN:                  "4012345678903",
			Category:             "Elektrowerkzeug",
			Unit:                 "Stück",
			StockCurrent:         4,
			StockReserved:        1,
			MinimumStock:         2,
			PurchasePriceNet:     79.90,
			SalesPriceGross:      129.99,
			SupplierNumber:       "SUPPL-003",
			DeliveryTimeInDays:   7,
			StorageLocation:      "Regal B3",
			WeightKg:             1.8,
			Dimensions:           "35x25x10",
			SerialNumberRequired: true,
			HazardClass:          "",
			Notes:                "Inklusive 2 Jahren Garantie",
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ArticleNumber:        "C4004",
			ShortName:            "Holzschrauben 4x30mm",
			LongName:             "Spanplattenschrauben 4x30mm verzinkt, 200 Stück",
			EAN:                  "4012345678904",
			Category:             "Schrauben",
			Unit:                 "Packung",
			StockCurrent:         25,
			StockReserved:        0,
			MinimumStock:         10,
			PurchasePriceNet:     3.75,
			SalesPriceGross:      6.49,
			SupplierNumber:       "SUPPL-004",
			DeliveryTimeInDays:   2,
			StorageLocation:      "Regal C4",
			WeightKg:             0.4,
			Dimensions:           "10x5x5",
			SerialNumberRequired: false,
			HazardClass:          "",
			Notes:                "",
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			ArticleNumber:        "D5005",
			ShortName:            "Farbe, weiß, 10L",
			LongName:             "Wandfarbe, reinweiß, 10 Liter",
			EAN:                  "4012345678905",
			Category:             "Farben",
			Unit:                 "Eimer",
			StockCurrent:         3,
			StockReserved:        1,
			MinimumStock:         5,
			PurchasePriceNet:     22.90,
			SalesPriceGross:      34.99,
			SupplierNumber:       "SUPPL-005",
			DeliveryTimeInDays:   4,
			StorageLocation:      "Regal D1",
			WeightKg:             10.5,
			Dimensions:           "30x30x40",
			SerialNumberRequired: false,
			HazardClass:          "niedrig",
			Notes:                "Mindesthaltbarkeit beachten",
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
	}

	// Artikel in der Datenbank speichern
	for _, article := range sampleArticles {
		err := r.articleRepo.Create(&article)
		if err != nil {
			return err
		}
	}

	return nil
}
