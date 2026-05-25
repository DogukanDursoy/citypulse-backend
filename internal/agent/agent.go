package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GeminiResponse, API'den dönen karmaşık JSON'u yakalamak için
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// YENİ: Fonksiyon artık 2 parametre alıyor: text ve imageBase64
func AnalyzeComplaint(text string, imageBase64 string) (string, error) {
	_ = godotenv.Load(".env")
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Modeli gemini-2.5-flash olarak bırakıyoruz (Multimodal işlemler için en hızlı ve doğrusu budur)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	// YENİ: Prompt'u Yapay Zekaya fotoğrafı da analiz etmesi gerektiğini söyleyecek şekilde güncelledik
	// YENİ: Yapay zekaya "Metin yoksa fotoğrafa bak" komutu eklenmiş prompt
// YENİ VE KUSURSUZ PROMPT MANTIĞI
	var promptText string
	if text == "" || strings.TrimSpace(text) == "" {
		// Vatandaş sadece fotoğraf attıysa Gemini'ye net talimat:
		promptText = "Sen uzman bir belediye yapay zekasısın. Vatandaş bir şikayet yazısı yazmadı, SADECE FOTOĞRAF GÖNDERDİ! Gözlerini aç ve bu fotoğraftaki problemi (çöp dağları, çukur, arıza vb.) tespit et. Yanıt formatın KESİNLİKLE sadece şu olsun: Kategori: [İsim], Öncelik: [Seviye], Birim: [Birim]"
	} else {
		// Hem yazı hem fotoğraf varsa:
		promptText = fmt.Sprintf("Sen uzman bir belediye yapay zekasısın. Vatandaşın şikayet metni: '%s'. Görsel gönderildiyse görseli de detaylıca incele. Yanıt formatın KESİNLİKLE sadece şu olsun: Kategori: [İsim], Öncelik: [Seviye], Birim: [Birim]", text)
	}	// YENİ: "Parts" (Parçalar) dizisini dinamik oluşturuyoruz
	var parts []interface{}

	// 1. Parça: Her zaman metni ekle
	parts = append(parts, map[string]interface{}{
		"text": promptText,
	})

	// 2. Parça: Eğer Flutter'dan fotoğraf geldiyse, onu da pakete ekle (Gemini'nin Gözleri açılıyor)
	if imageBase64 != "" {
		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]string{
				"mimeType": "image/jpeg", // Standart resim formatı
				"data":     imageBase64,
			},
		})
	}

	// Payload'u birleştir
	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": parts,
			},
		},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Eğer HTTP hata kodu döndüyse (400, 429 kota hatası, 500 vb.)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API Hatası (Kod %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("JSON Parse Hatası: %v", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("Yanıt boş döndü. Ham veri: %s", string(body))
}
