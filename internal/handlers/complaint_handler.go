package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"backendGo/internal/models"
	"backendGo/internal/repository"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GenerateTrackingCode() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 5)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return "CP-" + string(b)
}

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
				complaint.UserText,
			}
			writer.Write(row)
		}
	}
}

// 2. Yönetim Paneli İstatistikleri Endpoint'i
func GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	total, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{})
	pending, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{"status": "Beklemede"})
	resolved, _ := repository.ComplaintCollection.CountDocuments(context.TODO(), bson.M{"status": "Çözüldü"})

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

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}) // en yeni önce
	cursor, err := repository.ComplaintCollection.Find(context.TODO(), filter, opts)

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

// 4. Takip Kodu Sorgulama Endpoint'i
func TrackComplaint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Sadece GET metodu desteklenir", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Takip kodu eksik", http.StatusBadRequest)
		return
	}

	var complaint models.Complaint
	err := repository.ComplaintCollection.FindOne(context.TODO(), bson.M{"tracking_code": code}).Decode(&complaint)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bu takip koduna ait sikayet bulunamadi."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"tracking_code": complaint.TrackingCode,
		"status":        complaint.Status,
		"category":      complaint.Category,
		"priority":      complaint.Priority,
		"department":    complaint.Department,
	})
}

// 5. Şikayet Durumu Güncelleme Endpoint'i
// PATCH /api/complaints/{id}/status
func UpdateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPatch {
		http.Error(w, "Sadece PATCH desteklenir", http.StatusMethodNotAllowed)
		return
	}

	// URL'den ID'yi çek: /api/complaints/abc123/status
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Geçersiz URL", http.StatusBadRequest)
		return
	}
	idStr := parts[3] // /api/complaints/{id}/status

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Geçersiz ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		http.Error(w, "Geçersiz body, 'status' alanı gerekli", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"Beklemede": true, "İşlemde": true, "Çözüldü": true}
	if !validStatuses[body.Status] {
		http.Error(w, "Geçersiz status. Beklemede, İşlemde veya Çözüldü olmalı", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": body.Status}}
	result, err := repository.ComplaintCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil || result.MatchedCount == 0 {
		http.Error(w, "Şikayet bulunamadı veya güncellenemedi", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Durum güncellendi", "status": body.Status})
}
