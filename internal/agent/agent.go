package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

func AnalyzeComplaint(text string) (string, error) {
	_ = godotenv.Load(".env")
	apiKey := os.Getenv("GEMINI_API_KEY")

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey) // En sade ve doğru JSON yapısı
	payload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{
						"text": fmt.Sprintf("Sen bir belediye asistanısın. Şikayeti analiz et: '%s'. Yanıt formatın sadece şu olsun: Kategori: [İsim], Öncelik: [Seviye], Birim: [Birim]", text),
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //ML

	body, _ := io.ReadAll(resp.Body)

	// Eğer HTTP hata kodu döndüyse (400, 404, 500 vb.)
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
