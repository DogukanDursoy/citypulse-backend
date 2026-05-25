package main

import (
	"backendGo/internal/agent"
	"backendGo/internal/notification"
	"backendGo/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/joho/godotenv"
)

// YENİ: Dışarıdan (Flutter'dan) gelecek verinin formatına 'Image' eklendi
type RequestBody struct {
	Text  string `json:"text"`
	Email string `json:"email"`
	Image string `json:"image"` // Flutter'dan gelen Base64 görsel verisi
}

// Dışarıya (Flutter'a) döneceğimiz verinin formatı
type ResponseBody struct {
	Analysis string `json:"analysis,omitempty"`
	Error    string `json:"error,omitempty"`
}

// /api/analyze adresine gelen istekleri karşılayan fonksiyon
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Tarayıcının gönderdiği ön-kontrol (OPTIONS)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	// 1. Sadece POST isteklerini kabul et (Güvenlik)
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST metodu desteklenir", http.StatusMethodNotAllowed)
		return
	}

	// 2. Gelen JSON verisini oku
	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
		return
	}

	fmt.Println("--- Yeni Şikayet Geldi ---")
	fmt.Println("Metin:", req.Text)
	if req.Image != "" {
		fmt.Println("📷 Fotoğraf eklentisi algılandı! (Base64 formatında)")
	}

	// 3. Agent'ı (Gemini'yi) çağır
	// YENİ: Artık fonksiyona hem metni (Text) hem de fotoğrafı (Image) yolluyoruz!
	analysis, err := agent.AnalyzeComplaint(req.Text, req.Image)

	if err != nil {
		fmt.Println("🔴 GEMINI PATLADI! Sebep:", err)

		// Hata varsa direkt dön, veritabanına hatalı kayıt atma
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponseBody{Error: err.Error()})
		return
	}

	// ---------------- YENİ EKLENEN DB KAYIT KISMI ----------------
	// Gemini'den gelen analizi (Kategori, Öncelik, Birim) parçala
	var category, priority, department string
	parts := strings.Split(analysis, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Kategori:") {
			category = strings.TrimSpace(strings.TrimPrefix(part, "Kategori:"))
		} else if strings.HasPrefix(part, "Öncelik:") {
			priority = strings.TrimSpace(strings.TrimPrefix(part, "Öncelik:"))
		} else if strings.HasPrefix(part, "Birim:") {
			department = strings.TrimSpace(strings.TrimPrefix(part, "Birim:"))
		}
	}

	// Şikayet modelini oluştur
	newComplaint := repository.Complaint{
		UserText:   req.Text,
		Category:   category,
		Priority:   priority,
		Department: department,
		Status:     "Beklemede",
		CreatedAt:  time.Now(),
		// ImageBase64: req.Image, -> İleride bunu repository'e ekleyebiliriz!
	}

	// MongoDB'ye fırlat
	_, dbErr := repository.ComplaintCollection.InsertOne(context.TODO(), newComplaint)
	if dbErr != nil {
		fmt.Println("🔴 MongoDB Kayıt Hatası:", dbErr)
	} else {
		fmt.Println("🟢 Şikayet MongoDB Atlas'a başarıyla kaydedildi!")
	}
	// --------------------------------------------------------------

	// 4. Başarılı sonucu Flutter'a gönder
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{Analysis: analysis})

	// ------------------ YENİ: ASENKRON BİLDİRİM MOTORU ------------------
	// Eğer Flutter'dan mail gelmediyse test etmek için kendi mail adresini de buraya default yazabilirsin.
	targetEmail := req.Email
	if targetEmail == "" {
		targetEmail = "dursoydogukan@gmail.com" // Testlerin için kendi gerçek mailini yazabilirsin aga
	}
	go notification.SendComplaintEmail(targetEmail, req.Text, category, department, priority)
	// --------------------------------------------------------------------
}

func getComplaintsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodGet {
		http.Error(w, "Sadece GET metodu desteklenir", http.StatusMethodNotAllowed)
		return
	}

	var complaints []repository.Complaint

	// Veri tabanından tüm şikayetleri çek (bson.M{} boş filtre demek, yani hepsini getir)
	cursor, err := repository.ComplaintCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	// Bulunan tüm sonuçları complaints listesinin içine aktar
	if err = cursor.All(context.TODO(), &complaints); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Eğer liste boşsa, null yerine boş dizi [] dönsün (Flutter'da hata çıkmasın diye)
	if complaints == nil {
		complaints = []repository.Complaint{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(complaints)
}

func main() {
	// Ortam değişkenlerini yükle ve HATA VARSA SÖYLE
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Uyarı: .env dosyası bulunamadı. Bulut ortam değişkenleri kullanılacak.")
	}

	repository.ConnectDB()

	// Endpoint'i tanımla
	http.HandleFunc("/api/analyze", analyzeHandler)
	http.HandleFunc("/api/complaints", getComplaintsHandler)

	// Sunucuyu başlat
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 CityPulse API ayağa kalktı! (Port: %s)\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
