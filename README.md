# CityPulse — Akıllı Şehir Şikayet Yönetim Sistemi

CityPulse, vatandaşların belediyeye şikayet iletmesini ve yöneticilerin bu şikayetleri takip edip yönetmesini sağlayan yapay zeka destekli bir mobil + backend uygulamasıdır.

## 🎬 Demo

[![CityPulse Demo]](https://www.youtube.com/watch?v=DBxWszMi3xg)

---

## Özellikler

### 🤖 Yapay Zeka Destekli Şikayet Analizi
- Vatandaşın yazdığı metni Gemini AI ile analiz eder
- Şikayeti otomatik olarak kategorize eder, öncelik ve ilgili birimi belirler

### 📷 Fotoğraflı Şikayet
- Vatandaş galeri veya kameradan fotoğraf ekleyebilir
- Gemini AI fotoğrafı da analiz ederek şikayeti yorumlar

### 📊 Yönetim Paneli (Dashboard)
- Toplam şikayet sayısı, acil şikayet sayısı
- Birimlere göre şikayet yükü dağılımı
- Son gelen şikayetlerin canlı log akışı

### 📥 Excel Rapor İndirme
- Tüm şikayetleri CSV formatında dışa aktarma
- Tarih damgalı otomatik dosya adı

### 📲 SMS Bildirimi
- Şikayet kaydedilince vatandaşa Twilio üzerinden SMS gönderilir
- SMS içeriğinde takip kodu yer alır

### 🔍 Şikayet Takibi
- Her şikayete otomatik benzersiz takip kodu üretilir (örn: `CP-X7K2M`)
- Vatandaş chat ekranına kodunu yazarak şikayetinin durumunu sorgulayabilir

### 📋 Geçmiş & Şikayet Yükü
- Drawer menüsünde tüm geçmiş şikayetler listelenir
- Dashboard'da birimlere göre şikayet yoğunluğu görüntülenir

### 🔐 Yönetici Kimlik Doğrulama
- Dashboard şifre korumalıdır
- Şifre backend ortam değişkeninden alınır, kodda yer almaz

### ✅ Şikayet Durum Yönetimi
- Yönetici her şikayetin durumunu değiştirebilir: `Beklemede` → `İşlemde` → `Çözüldü`
- Değişiklik anında MongoDB'ye yansır

### 📍 GPS Konum Etiketleme
- Şikayet gönderilirken kullanıcının GPS koordinatları kaydedilir
- Dashboard'da konumu olan şikayete tıklanınca Google Maps'te görüntülenir

---

## Teknolojiler

| Katman | Teknoloji |
|--------|-----------|
| Mobil | Flutter |
| Backend | Go |
| Veritabanı | MongoDB Atlas |
| Yapay Zeka | Google Gemini API |
| SMS | Twilio |
| Deployment | Render |

---

## Kurulum

### Gerekli Ortam Değişkenleri (.env)

```
MONGO_URI=...
GEMINI_API_KEY=...
TWILIO_ACCOUNT_SID=...
TWILIO_AUTH_TOKEN=...
TWILIO_FROM_PHONE=...
ADMIN_PASSWORD=...
```

### Backend

```bash
cd backendGo
go run ./cmd/main.go
```

### Flutter

```bash
cd frontendFlutter
flutter pub get
flutter run
```
