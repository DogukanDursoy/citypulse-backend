package notification

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendSMS, Twilio REST API üzerinden asenkron SMS gönderir.
func SendSMS(toPhone, category, priority string) {
	fmt.Println("📲 [ADIM 1] Asenkron SMS kuyruğu tetiklendi...")

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromPhone := os.Getenv("TWILIO_FROM_PHONE") // Twilio'nun sana vereceği numara

	if accountSid == "" || authToken == "" || fromPhone == "" {
		fmt.Println("⚠️ [HATA] Twilio ayarları eksik! SMS pas geçildi.")
		return
	}

	// Twilio'nun mesaj gönderme uç noktası (HTTP POST)
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	// SMS İçeriği (Türkçe karakter kullanmamak SMS'te her zaman daha garantidir)
	bodyText := fmt.Sprintf("CityPulse test onayi. Sistem aktif.")

	// API'ye gidecek verileri paketliyoruz
	msgData := url.Values{}
	msgData.Set("To", toPhone)
	msgData.Set("From", fromPhone)
	msgData.Set("Body", bodyText)

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(msgData.Encode()))
	if err != nil {
		fmt.Println("🔴 [HATA] İstek oluşturulamadı:", err)
		return
	}

	// Twilio kimlik doğrulaması
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	fmt.Println("⏳ [ADIM 2] Twilio API'sine istek atılıyor (Render'ı delip geçiyoruz)...")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("🔴 [HATA] API'ye ulaşılamadı:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("🟢 [BAŞARI] SMS vatandaşa başarıyla uçuruldu!")
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("🔴 [HATA] Twilio'dan 400 yedik! Sebebi: %s\n", string(bodyBytes))
	}
}
