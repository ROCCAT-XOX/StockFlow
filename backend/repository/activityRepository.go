// backend/repository/activityRepository.go (angepasste Version)
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

// ActivityRepository enthält alle Datenbankoperationen für das Activity-Modell
type ActivityRepository struct {
	collection *mongo.Collection
}

// NewActivityRepository erstellt ein neues ActivityRepository
func NewActivityRepository() *ActivityRepository {
	return &ActivityRepository{
		collection: db.GetCollection("activities"),
	}
}

// Create erstellt eine neue Aktivität
func (r *ActivityRepository) Create(activity *model.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		return err
	}

	activity.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindRecent findet die neuesten Aktivitäten, begrenzt durch limit
func (r *ActivityRepository) FindRecent(limit int) ([]*model.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung und Begrenzung
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}). // Absteigend nach Zeitstempel
		SetLimit(int64(limit))

	var activities []*model.Activity
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var activity model.Activity
		if err := cursor.Decode(&activity); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// FindByUserID findet Aktivitäten eines bestimmten Benutzers
func (r *ActivityRepository) FindByUserID(userID string, limit int) ([]*model.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ID in ObjectID umwandeln
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Optionen für die Sortierung und Begrenzung
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}). // Absteigend nach Zeitstempel
		SetLimit(int64(limit))

	var activities []*model.Activity
	cursor, err := r.collection.Find(ctx, bson.M{"userId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var activity model.Activity
		if err := cursor.Decode(&activity); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// FindByTargetID findet Aktivitäten für ein bestimmtes Zielobjekt
func (r *ActivityRepository) FindByTargetID(targetID string, limit int) ([]*model.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ID in ObjectID umwandeln
	objID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return nil, err
	}

	// Optionen für die Sortierung und Begrenzung
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}). // Absteigend nach Zeitstempel
		SetLimit(int64(limit))

	var activities []*model.Activity
	cursor, err := r.collection.Find(ctx, bson.M{"targetId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var activity model.Activity
		if err := cursor.Decode(&activity); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// LogActivity fügt eine neue Aktivität hinzu und gibt die erstellte Aktivität zurück
func (r *ActivityRepository) LogActivity(
	activityType model.ActivityType,
	userId primitive.ObjectID,
	userName string,
	targetId primitive.ObjectID,
	targetType, targetName, description string,
	quantity float64,
) (*model.Activity, error) {
	activity := &model.Activity{
		Type:        activityType,
		UserID:      userId,
		UserName:    userName,
		TargetID:    targetId,
		TargetType:  targetType,
		TargetName:  targetName,
		Description: description,
		Quantity:    quantity,
		Timestamp:   time.Now(),
	}

	err := r.Create(activity)
	if err != nil {
		return nil, err
	}

	return activity, nil
}

// CountActivitiesSince zählt die Anzahl der Aktivitäten seit einem bestimmten Zeitpunkt
func (r *ActivityRepository) CountActivitiesSince(since time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"timestamp": bson.M{"$gte": since},
	}

	return r.collection.CountDocuments(ctx, filter)
}

// GetActivityCountByType gibt die Anzahl der Aktivitäten pro Typ zurück
func (r *ActivityRepository) GetActivityCountByType() (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für die Anzahl pro Typ
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ergebnisse in Map umwandeln
	countByType := make(map[string]int64)
	for _, result := range results {
		countByType[result.ID] = result.Count
	}

	return countByType, nil
}

// Delete löscht eine Aktivität
func (r *ActivityRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// DeleteByTargetID löscht alle Aktivitäten für ein bestimmtes Zielobjekt
func (r *ActivityRepository) DeleteByTargetID(targetID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"targetId": objID})
	return err
}
