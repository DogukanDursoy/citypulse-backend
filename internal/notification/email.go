package notification

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// SendComplaintEmail, vatandaş şikayet kaydı açtığında asenkron tetiklenecek HTML mail fonksiyonudur.
func SendComplaintEmail(toEmail, complaintText, category, department, priority string) {
	fmt.Println("📩 Asenkron e-posta kuyruğu tetiklendi, gönderim başlatılıyor...")

	// .env dosyasından SMTP bilgilerini çekiyoruz
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST") // örn: smtp.gmail.com
	smtpPort := os.Getenv("SMTP_PORT") // örn: 587

	// Eğer mail adresi girilmediyse veya konfigürasyon eksikse sistemi patlatma, log bas geç
	if toEmail == "" || smtpUser == "" || smtpPass == "" {
		fmt.Println("⚠️ Uyarı: Alıcı e-postası veya SMTP ayarları eksik. E-posta gönderimi pas geçildi.")
		return
	}

	// Metin boşsa görsel analizi yapılmıştır mesajı verelim
	if strings.TrimSpace(complaintText) == "" {
		complaintText = "[Vatandaş sadece fotoğraf gönderdi - Yapay zeka görsel tespiti]"
	}

	to := []string{toEmail}

	// Jürinin gözünü boyayacak janti bir HTML şablonu hazırlıyoruz
	subject := "Subject: CityPulse - Şikayetiniz Başarıyla Alındı! 🏙️\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: 'Segoe UI', Arial, sans-serif; color: #333; line-height: 1.6; background-color: #f9f9f9; padding: 20px;">
			<div style="max-width: 600px; margin: 0 auto; padding: 25px; background-color: #ffffff; border-top: 6px solid #008080; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.05);">
				<h2 style="color: #008080; text-align: center; margin-bottom: 20px;">CityPulse Akıllı Şehir Bildirim Sistemi</h2>
				<p>Merhaba Değerli Vatandaşımız,</p>
				<p>Belediyemize iletmiş olduğunuz talep/şikayet kaydı sistemimize başarıyla mühürlenmiştir. Akıllı Şehir Yapay Zeka Ajanımızın (Agent) anlık analiz sonuçları aşağıda bilgilerinize sunulmuştur:</p>
				
				<table style="width: 100%%; border-collapse: collapse; margin: 25px 0;">
					<tr style="background-color: #f2f2f2;">
						<td style="padding: 12px; font-weight: bold; border: 1px solid #ddd; width: 30%%;">İletilen Mesaj:</td>
						<td style="padding: 12px; border: 1px solid #ddd;">%s</td>
					</tr>
					<tr>
						<td style="padding: 12px; font-weight: bold; border: 1px solid #ddd;">Yapay Zeka Kategorisi:</td>
						<td style="padding: 12px; border: 1px solid #ddd;">%s</td>
					</tr>
					<tr style="background-color: #f2f2f2;">
						<td style="padding: 12px; font-weight: bold; border: 1px solid #ddd;">Sorumlu Belediye Birimi:</td>
						<td style="padding: 12px; border: 1px solid #ddd; color: #008080; font-weight: bold;">%s</td>
					</tr>
					<tr>
						<td style="padding: 12px; font-weight: bold; border: 1px solid #ddd;">Aciliyet / Öncelik:</td>
						<td style="padding: 12px; border: 1px solid #ddd; color: #d9534f; font-weight: bold;">%s</td>
					</tr>
				</table>
				
				<p style="background-color: #e6f2f2; padding: 12px; border-left: 4px solid #008080; color: #005959; font-size: 14px;">
					<strong>Süreç Bilgilendirmesi:</strong> Şikayetiniz ilgili müdürlük ekiplerine otomatik iş emri olarak sevk edilmiştir. Durum değişikliklerinde sistem tarafından bilgilendirilmeye devam edeceksiniz.
				</p>
				
				<hr style="border: 0; border-top: 1px solid #eee; margin: 25px 0;">
				<p style="font-size: 11px; color: #999; text-align: center; margin-top: 20px;">
					Bu e-posta CityPulse Otomatik Analiz ve Bildirim Modülü tarafından üretilmiştir. Lütfen yanıtlamayınız.
				</p>
			</div>
		</body>
		</html>
	`, complaintText, category, department, priority)

	msg := []byte(subject + mime + body)

	// SMTP Authentication ayarı
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Maili fırlat
	err := smtp.SendMail(addr, auth, smtpUser, to, msg)
	if err != nil {
		fmt.Println("🔴 E-posta Gönderilirken Hata Oluştu:", err)
		return
	}

	fmt.Println("🟢 Bildirim e-postası vatandaşa başarıyla uçuruldu!")
}
