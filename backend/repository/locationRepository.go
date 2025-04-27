// backend/repository/locationRepository.go
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

// LocationRepository enthält alle Datenbankoperationen für das Location-Modell
type LocationRepository struct {
	collection *mongo.Collection
}

// NewLocationRepository erstellt ein neues LocationRepository
func NewLocationRepository() *LocationRepository {
	return &LocationRepository{
		collection: db.GetCollection("locations"),
	}
}

// Create erstellt einen neuen Lagerort
func (r *LocationRepository) Create(location *model.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Standardwerte setzen
	if location.CreatedAt.IsZero() {
		location.CreatedAt = time.Now()
	}
	if location.UpdatedAt.IsZero() {
		location.UpdatedAt = time.Now()
	}
	if !location.IsActive {
		location.IsActive = true
	}

	result, err := r.collection.InsertOne(ctx, location)
	if err != nil {
		return err
	}

	location.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet einen Lagerort anhand seiner ID
func (r *LocationRepository) FindByID(id string) (*model.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var location model.Location
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&location)
	if err != nil {
		return nil, err
	}

	return &location, nil
}

// FindAll findet alle Lagerorte
func (r *LocationRepository) FindAll() ([]*model.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Name
	opts := options.Find().SetSort(bson.D{
		{Key: "name", Value: 1},
	})

	var locations []*model.Location
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var location model.Location
		if err := cursor.Decode(&location); err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

// FindWarehouses findet alle Hauptlager (ohne Parent)
func (r *LocationRepository) FindWarehouses() ([]*model.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Name
	opts := options.Find().SetSort(bson.D{
		{Key: "name", Value: 1},
	})

	var locations []*model.Location
	cursor, err := r.collection.Find(ctx, bson.M{"type": model.LocationTypeWarehouse}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var location model.Location
		if err := cursor.Decode(&location); err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

// FindByParentID findet alle Lagerorte mit einem bestimmten übergeordneten Lagerort
func (r *LocationRepository) FindByParentID(parentID string) ([]*model.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		return nil, err
	}

	// Optionen für die Sortierung nach Typ und Name
	opts := options.Find().SetSort(bson.D{
		{Key: "type", Value: 1},
		{Key: "name", Value: 1},
	})

	var locations []*model.Location
	cursor, err := r.collection.Find(ctx, bson.M{"parentId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var location model.Location
		if err := cursor.Decode(&location); err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}

// Update aktualisiert einen bestehenden Lagerort
func (r *LocationRepository) Update(location *model.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// UpdatedAt-Zeitstempel aktualisieren
	location.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": location.ID},
		bson.M{"$set": location},
	)
	return err
}

// Delete löscht einen Lagerort
func (r *LocationRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Prüfen, ob Unterkategorien existieren
	count, err := r.collection.CountDocuments(ctx, bson.M{"parentId": objID})
	if err != nil {
		return err
	}
	if count > 0 {
		return mongo.ErrNoDocuments // Fehlertyp hier nicht optimal, aber einfach zu erkennen
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// BuildLocationTree erstellt einen hierarchischen Baum aus Lagerorten
func (r *LocationRepository) BuildLocationTree() (map[primitive.ObjectID]*model.Location, error) {
	// Alle Lagerorte laden
	locations, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	// Map für schnellen Zugriff erstellen
	locationMap := make(map[primitive.ObjectID]*model.Location)
	for _, loc := range locations {
		locationMap[loc.ID] = loc
	}

	return locationMap, nil
}

// GetLocationPath gibt den vollständigen Pfad für einen Lagerort zurück
func (r *LocationRepository) GetLocationPath(locationID string) (string, error) {
	// Lagerort finden
	location, err := r.FindByID(locationID)
	if err != nil {
		return "", err
	}

	// Alle Lagerorte laden und Map erstellen
	locationMap, err := r.BuildLocationTree()
	if err != nil {
		return "", err
	}

	// Pfad generieren
	return location.GetFullPath(locationMap), nil
}
