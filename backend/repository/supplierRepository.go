// backend/repository/supplierRepository.go
package repository

import (
	"context"
	"time"

	"StockFlow/backend/db"
	"StockFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SupplierRepository enthält alle Datenbankoperationen für das Supplier-Modell
type SupplierRepository struct {
	collection *mongo.Collection
}

// NewSupplierRepository erstellt ein neues SupplierRepository
func NewSupplierRepository() *SupplierRepository {
	return &SupplierRepository{
		collection: db.GetCollection("suppliers"),
	}
}

// Create erstellt einen neuen Lieferanten
func (r *SupplierRepository) Create(supplier *model.Supplier) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob bereits ein Lieferant mit diesem Code existiert
	if supplier.SupplierCode != "" {
		count, err := r.collection.CountDocuments(ctx, bson.M{"supplierCode": supplier.SupplierCode})
		if err != nil {
			return err
		}
		if count > 0 {
			return mongo.ErrNoDocuments // Fehlertyp hier nicht optimal, aber einfach zu erkennen
		}
	}

	// Standardwerte setzen
	if supplier.CreatedAt.IsZero() {
		supplier.CreatedAt = time.Now()
	}
	if supplier.UpdatedAt.IsZero() {
		supplier.UpdatedAt = time.Now()
	}
	if !supplier.IsActive {
		supplier.IsActive = true
	}

	result, err := r.collection.InsertOne(ctx, supplier)
	if err != nil {
		return err
	}

	supplier.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet einen Lieferanten anhand seiner ID
func (r *SupplierRepository) FindByID(id string) (*model.Supplier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var supplier model.Supplier
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&supplier)
	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

// FindBySupplierCode findet einen Lieferanten anhand seines Codes
func (r *SupplierRepository) FindBySupplierCode(code string) (*model.Supplier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var supplier model.Supplier
	err := r.collection.FindOne(ctx, bson.M{"supplierCode": code}).Decode(&supplier)
	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

// FindAll findet alle Lieferanten
func (r *SupplierRepository) FindAll() ([]*model.Supplier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Name
	opts := options.Find().SetSort(bson.D{
		{Key: "name", Value: 1},
	})

	var suppliers []*model.Supplier
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var supplier model.Supplier
		if err := cursor.Decode(&supplier); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, &supplier)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

// FindActive findet alle aktiven Lieferanten
func (r *SupplierRepository) FindActive() ([]*model.Supplier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Name
	opts := options.Find().SetSort(bson.D{
		{Key: "name", Value: 1},
	})

	var suppliers []*model.Supplier
	cursor, err := r.collection.Find(ctx, bson.M{"isActive": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var supplier model.Supplier
		if err := cursor.Decode(&supplier); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, &supplier)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

// Update aktualisiert einen bestehenden Lieferanten
func (r *SupplierRepository) Update(supplier *model.Supplier) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// UpdatedAt-Zeitstempel aktualisieren
	supplier.UpdatedAt = time.Now()

	// Prüfen, ob bereits ein anderer Lieferant mit diesem Code existiert
	if supplier.SupplierCode != "" {
		var existingSupplier model.Supplier
		err := r.collection.FindOne(ctx, bson.M{
			"supplierCode": supplier.SupplierCode,
			"_id":          bson.M{"$ne": supplier.ID},
		}).Decode(&existingSupplier)

		// Wenn ein Dokument gefunden wurde, bedeutet das, dass der Code bereits verwendet wird
		if err == nil {
			return mongo.ErrNoDocuments
		}
		// Wenn der Fehler nicht "nicht gefunden" ist, ist es ein anderer Fehler
		if err != mongo.ErrNoDocuments {
			return err
		}
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": supplier.ID},
		bson.M{"$set": supplier},
	)
	return err
}

// Delete löscht einen Lieferanten
func (r *SupplierRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Search sucht Lieferanten anhand verschiedener Kriterien
func (r *SupplierRepository) Search(query string) ([]*model.Supplier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Suche in mehreren Feldern
	filter := bson.M{
		"$or": []bson.M{
			{"supplierCode": bson.M{"$regex": query, "$options": "i"}},
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"contactPerson": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	var suppliers []*model.Supplier
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var supplier model.Supplier
		if err := cursor.Decode(&supplier); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, &supplier)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

// Count zählt die Gesamtzahl der Lieferanten
func (r *SupplierRepository) Count() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountActive zählt die Anzahl der aktiven Lieferanten
func (r *SupplierRepository) CountActive() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"isActive": true})
}
