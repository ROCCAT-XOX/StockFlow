// backend/model/activity.go (überarbeitete Version)
package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ActivityType repräsentiert den Typ einer Aktivität
type ActivityType string

const (
	// Artikel-bezogene Aktivitäten
	ActivityTypeArticleAdded   ActivityType = "article_added"
	ActivityTypeArticleUpdated ActivityType = "article_updated"
	ActivityTypeArticleDeleted ActivityType = "article_deleted"
	ActivityTypeStockAdjusted  ActivityType = "stock_adjusted" // Neu: Für Bestandsanpassungen
	ActivityTypeStockTaking    ActivityType = "stock_taking"   // Neu: Für Inventur

	// System-bezogene Aktivitäten
	ActivityTypeUserAdded       ActivityType = "user_added"
	ActivityTypeUserUpdated     ActivityType = "user_updated"
	ActivityTypeUserDeleted     ActivityType = "user_deleted"
	ActivityTypeUserLogin       ActivityType = "user_login" // Neu: Für Login-Protokollierung
	ActivityTypeSupplierAdded   ActivityType = "supplier_added"
	ActivityTypeSupplierUpdated ActivityType = "supplier_updated"
	ActivityTypeSupplierDeleted ActivityType = "supplier_deleted"
)

// Activity repräsentiert eine Aktivität im System
type Activity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        ActivityType       `bson:"type" json:"type"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserName    string             `bson:"userName" json:"userName"`
	TargetID    primitive.ObjectID `bson:"targetId,omitempty" json:"targetId,omitempty"`
	TargetType  string             `bson:"targetType" json:"targetType"`
	TargetName  string             `bson:"targetName" json:"targetName"`
	Description string             `bson:"description" json:"description"`
	Quantity    float64            `bson:"quantity,omitempty" json:"quantity,omitempty"` // Neu: Für Mengenangaben bei Bestandsänderungen
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}

// GetIconClass gibt die CSS-Klasse für das Icon basierend auf dem Aktivitätstyp zurück
func (a *Activity) GetIconClass() string {
	switch a.Type {
	case ActivityTypeArticleAdded, ActivityTypeStockAdjusted:
		return "bg-green-500"
	case ActivityTypeArticleUpdated, ActivityTypeStockTaking:
		return "bg-blue-500"
	case ActivityTypeArticleDeleted:
		return "bg-red-500"
	case ActivityTypeUserLogin:
		return "bg-yellow-500"
	default:
		return "bg-gray-500"
	}
}

// GetIconSVG gibt das SVG-Icon basierend auf dem Aktivitätstyp zurück
func (a *Activity) GetIconSVG() string {
	switch a.Type {
	case ActivityTypeArticleAdded:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM14 11a1 1 0 011 1v1h1a1 1 0 110 2h-1v1a1 1 0 11-2 0v-1h-1a1 1 0 110-2h1v-1a1 1 0 011-1z\" /></svg>"
	case ActivityTypeArticleUpdated:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z\" /></svg>"
	case ActivityTypeArticleDeleted:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z\" clip-rule=\"evenodd\" /></svg>"
	case ActivityTypeStockAdjusted:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M10 3a1 1 0 01.707.293l3 3a1 1 0 01-1.414 1.414L10 5.414 7.707 7.707a1 1 0 01-1.414-1.414l3-3A1 1 0 0110 3zm-3.707 9.293a1 1 0 011.414 0L10 14.586l2.293-2.293a1 1 0 011.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z\" clip-rule=\"evenodd\" /></svg>"
	case ActivityTypeStockTaking:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M9 2a1 1 0 000 2h2a1 1 0 100-2H9z\" /><path fill-rule=\"evenodd\" d=\"M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z\" clip-rule=\"evenodd\" /></svg>"
	case ActivityTypeUserLogin:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M3 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm7.707 3.293a1 1 0 010 1.414L9.414 9H17a1 1 0 110 2H9.414l1.293 1.293a1 1 0 01-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0z\" clip-rule=\"evenodd\" /></svg>"
	default:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z\" clip-rule=\"evenodd\" /></svg>"
	}
}

// FormatTimeAgo formatiert den Zeitstempel als "vor X Zeit" (z.B. "vor 5 Minuten")
func (a *Activity) FormatTimeAgo() string {
	now := time.Now()
	diff := now.Sub(a.Timestamp)

	if diff < time.Minute {
		return "gerade eben"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("vor %d Minute%s", minutes, pluralS(minutes))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("vor %d Stunde%s", hours, pluralS(hours))
	} else if diff < 48*time.Hour {
		return "gestern"
	} else {
		return a.Timestamp.Format("02.01.2006 15:04")
	}
}

// Hilfsfunktion für Plural-Endungen
func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "n"
}
