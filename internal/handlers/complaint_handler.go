package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"backendGo/internal/models"
	"backendGo/internal/repository"
)

// 1. Dışa Aktarma (Excel/CSV) Endpoint'i
func ExportComplaintsCSV(w http.ResponseWriter, r *http.Request) {
	fileName := fmt.Sprintf("CityPulse_Rapor_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	headers := []string{"ID", "Kategori", "Durum", "Öncelik", "Departman", "Şikayet Metni"}
	if err := writer.Write(headers); err != nil {
		http.Error(w, "CSV oluşturulamadı", http.StatusInternalServerError)
		return
	}

	cursor, err := repository.ComplaintCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Veriler çekilemedi", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var complaint models.Complaint
		if err := cursor.Decode(&complaint); err == nil {
			row := []string{
				complaint.ID.Hex(),
				complaint.Category,
				complaint.Status,
				complaint.Priority,
				complaint.Department,
				complaint.Text,
			}
			writer.Write(row)
		}
	}
}

// 2. Yönetim Paneli İstatistikleri Endpoint'i
func GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	total, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{})
	pending, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{"status": "Pending"})
	resolved, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{"status": "Resolved"})

	response := map[string]int64{
		"total":    total,
		"pending":  pending,
		"resolved": resolved,
	}
	json.NewEncoder(w).Encode(response)
}

// 3. Şikayetleri Listeleme ve Filtreleme Endpoint'i
func GetComplaints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filter := bson.M{}
	category := r.URL.Query().Get("category")
	if category != "" {
		filter["category"] = category
	}

	cursor, err := repository.ComplaintCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Veri çekilemedi", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var complaints []models.Complaint
	if err = cursor.All(context.TODO(), &complaints); err != nil {
		http.Error(w, "Veriler dönüştürülemedi", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(complaints)
}