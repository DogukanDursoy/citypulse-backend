import 'dart:convert';
import 'dart:io'; // YENİ: Dosya işlemleri için
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart'; // YENİ: Fotoğraf seçmek için
import 'package:url_launcher/url_launcher.dart';
void main() {
  WidgetsFlutterBinding.ensureInitialized();
  SystemChrome.setEnabledSystemUIMode(SystemUiMode.immersiveSticky);
  runApp(const CityPulseApp());
}
// ============================================================================
// AÇILIŞ EKRANI (SPLASH SCREEN)
// ============================================================================
class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}


class _SplashScreenState extends State<SplashScreen> {
  @override
  void initState() {
    super.initState();
    // 2.5 saniye ekranda kalıp sonra ChatScreen'e geçer
    Future.delayed(const Duration(milliseconds: 2500), () {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const ChatScreen()),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.teal, // Arka plan ana rengimiz
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // İkon
            const Icon(
              Icons.location_city,
              size: 80,
              color: Colors.white,
            ),
            const SizedBox(height: 16),
            // Yazı
            const Text(
              'CityPulse',
              style: TextStyle(
                fontSize: 42,
                fontWeight: FontWeight.bold,
                color: Colors.white,
                letterSpacing: 2.0, // Harfler arası biraz boşluk kurumsal durur
              ),
            ),
            const SizedBox(height: 8),
            // Alt başlık
            Text(
              'Akıllı Şehir Asistanı',
              style: TextStyle(
                fontSize: 16,
                color: Colors.teal.shade100,
                fontWeight: FontWeight.w500,
              ),
            ),
            const SizedBox(height: 50),
            // Ufak bir yükleniyor animasyonu
            const CircularProgressIndicator(
              color: Colors.white,
              strokeWidth: 3,
            )
          ],
        ),
      ),
    );
  }
}

class CityPulseApp extends StatelessWidget {
  const CityPulseApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'CityPulse',
      debugShowCheckedModeBanner:
          false, // Sağ üstteki o çirkin DEBUG yazısını kaldırır
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal),
        useMaterial3: true,
      ),
      home: const SplashScreen(),
    );
  }
}

// 1. Yeni Veri Yapımız: Mesaj Modeli (Fotoğraf desteği eklendi)
class ChatMessage {
  final String text;
  final bool
      isUser; // True ise sağda (Biz), False ise solda (Yapay Zeka) çıkacak
  final String?
      base64Image; // YENİ: Kullanıcı fotoğraf attıysa balon içinde göstereceğiz

  ChatMessage({required this.text, required this.isUser, this.base64Image});
}

class ChatScreen extends StatefulWidget {
  const ChatScreen({super.key});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final TextEditingController _controller = TextEditingController();
  final ImagePicker _picker = ImagePicker(); // YENİ: Kamera/Galeri nesnesi

  String _userName = "";
  bool _isWaitingForName =
      true; // Uygulama ilk açıldığında isim bekliyor olacak

  // Chat mesajlarını tutan liste
  final List<ChatMessage> _messages = [];
  bool _isLoading = false;

  // Geçmiş şikayetleri tutan liste ve yüklenme durumu
  List<dynamic> _complaintHistory = [];
  bool _isLoadingHistory = false;

  File? _selectedImage; // YENİ: Seçilen fotoğraf
  String? _base64ImageString; // YENİ: Fotoğrafın backend'e gidecek string hali

  @override
  void initState() {
    super.initState();
    // Uygulama açılır açılmaz veritabanından geçmişi çek
    _fetchHistory();

    _messages.add(ChatMessage(
        text:
            'CityPulse Şikayet Takip Sistemine hoş geldiniz! Size nasıl hitap edebilirim?',
        isUser: false));
  }

  // YENİ: Fotoğraf Seçme Fonksiyonu
  Future<void> _pickImage(ImageSource source) async {
    try {
      final XFile? image =
          await _picker.pickImage(source: source, imageQuality: 50);
      if (image != null) {
        final bytes = await image.readAsBytes();
        setState(() {
          _selectedImage = File(image.path);
          _base64ImageString = base64Encode(bytes); // Base64'e çevirip sakla
        });
      }
    } catch (e) {
      print("Fotoğraf seçilirken hata oluştu: $e");
    }
  }

  // MongoDB'den Geçmiş Şikayetleri Çekme Fonksiyonu
  Future<void> _fetchHistory() async {
    setState(() {
      _isLoadingHistory = true;
    });

    try {
      final url = Uri.parse(
          'https://citypulse-backend-wjt6.onrender.com/api/complaints');
      final response = await http.get(url);

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        setState(() {
          // En yeni şikayet en üstte görünsün diye listeyi tersine çeviriyoruz
          _complaintHistory = List.from(data.reversed);
        });
      } else {
        print('Geçmiş çekilemedi. Hata kodu: ${response.statusCode}');
      }
    } catch (e) {
      print('Geçmiş çekilirken hata oluştu: $e');
    } finally {
      setState(() {
        _isLoadingHistory = false;
      });
    }
  }

  Future<void> _sendMessage() async {
    // Mesaj kutusu boşsa VE fotoğraf da seçilmediyse hiçbir şey yapma (Değiştirildi)
    if (_controller.text.trim().isEmpty && _selectedImage == null) return;

    final userText = _controller.text;
    final sendingImage = _base64ImageString; // Anlık fotoğrafı al

    _controller.clear(); // Mesajı yollayınca kutuyu hemen temizle
    setState(() {
      _selectedImage = null; // Ekrandan fotoğraf ön izlemesini kaldır
      _base64ImageString = null;
    });

    // 1. Kullanıcının mesajını (ve varsa fotoğrafını) ekrana (sağa) ekle
    setState(() {
      _messages.add(
          ChatMessage(text: userText, isUser: true, base64Image: sendingImage));
    });

    // --- 1. SENARYO: EĞER SİSTEM İSİM BEKLİYORSA ---
    if (_isWaitingForName) {
      setState(() {
        _userName = userText; // Kullanıcının yazdığını isim olarak kaydet
        _isWaitingForName = false; // İsim alma evresini kapat

        // İsme özel şık bir karşılama yap
        _messages.add(ChatMessage(
            text:
                'Memnun oldum $_userName! Bugün şehrimizde gördüğün bir problemi fotoğraf ekleyerek veya yazarak benimle paylaşabilirsin.',
            isUser: false));
      });
      return; // İşlem bitti, backend'e boşuna gitmemek için burada kes!
    }

    // --- 2. SENARYO: KONUŞMAYI BİTİRME (Hayır/Yok Kontrolü) ---
    final lowerText = userText.toLowerCase().trim();

    if (lowerText == 'hayır' ||
        lowerText == 'yok' ||
        lowerText == 'hayır yok' ||
        lowerText == 'teşekkürler' ||
        lowerText == 'sağol') {
      setState(() {
        _messages.add(ChatMessage(
            text:
                'Ben teşekkür ederim $_userName! Başka bir sorun olduğunda CityPulse asistanı olarak her zaman buradayım. İyi günler dilerim! 🏙️',
            isUser: false));
      });
      return;
    }

    // --- 3. SENARYO: ŞİKAYET GÖNDERİLİYORSA (Normal Akış) ---
    setState(() {
      _isLoading = true;
    });

    try {
      final url =
          Uri.parse('https://citypulse-backend-wjt6.onrender.com/api/analyze');

      // YENİ PAYLOAD: Artık 'image' alanını da gönderiyoruz
      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'text': userText,
          'image': sendingImage ?? "", // Fotoğraf yoksa boş string gidecek
        }),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final analysis = data['analysis'] ?? 'Sonuç alınamadı.';

        setState(() {
          // 1. Gemini'nin Analizi
          _messages.add(ChatMessage(text: analysis, isUser: false));

          // 2. Ardından gelen o can alıcı soru
          _messages.add(ChatMessage(
              text: 'Başka bildirmek istediğiniz bir sorun var mı $_userName?',
              isUser: false));
        });

        // Şikayet başarıyla iletildiyse Drawer (Geçmiş) menüsünü güncelle
        _fetchHistory();
      } else {
        setState(() {
          _messages.add(ChatMessage(
              text: 'Sunucu Hatası: ${response.statusCode}', isUser: false));
        });
      }
    } catch (e) {
      setState(() {
        _messages.add(ChatMessage(
            text: 'Bağlantı Hatası: Go sunucusu açık mı?\n$e', isUser: false));
      });
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade50,
      appBar: AppBar(
        title: const Text('CityPulse Asistan',
            style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        elevation: 1,
      ),

      // DİNAMİK YANDAN AÇILAN MENÜ (DRAWER)
      drawer: Drawer(
        backgroundColor: Colors.white, // Arkaplan saydamlığını kapatır!
        child: Column(
          children: [
            const DrawerHeader(
              decoration: BoxDecoration(color: Colors.teal),
              margin: EdgeInsets.zero,
              child: SizedBox(
                width: double.infinity,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    Icon(Icons.location_city, color: Colors.white, size: 40),
                    SizedBox(height: 10),
                    Text('CityPulse Menü',
                        style: TextStyle(color: Colors.white, fontSize: 22)),
                  ],
                ),
              ),
            ),

            // ================= DASHBOARD BUTONU =================
            ListTile(
              leading:
                  const Icon(Icons.dashboard, color: Colors.teal, size: 28),
              title: const Text(
                'Yönetim Paneli (Dashboard)',
                style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: Colors.teal,
                    fontSize: 15),
              ),
              subtitle: const Text('Belediye istatistikleri ve analizler'),
              onTap: () {
                Navigator.pop(context); // Önce menüyü kapat
                Navigator.push(
                  context,
                  MaterialPageRoute(
                      builder: (context) => const DashboardScreen()),
                );
              },
            ),
            const Divider(height: 1, thickness: 1),
            // ====================================================

            Padding(
              padding:
                  const EdgeInsets.only(top: 12.0, left: 16.0, bottom: 8.0),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Geçmiş Şikayetler (${_complaintHistory.length})',
                  style: TextStyle(
                      color: Colors.grey.shade600, fontWeight: FontWeight.bold),
                ),
              ),
            ),

            // Burası veritabanından dinamik doluyor
            Expanded(
              child: _isLoadingHistory
                  ? const Center(child: CircularProgressIndicator())
                  : _complaintHistory.isEmpty
                      ? const Center(child: Text('Henüz şikayet bulunmuyor.'))
                      : ListView.builder(
                          padding: EdgeInsets.zero,
                          itemCount: _complaintHistory.length,
                          itemBuilder: (context, index) {
                            final item = _complaintHistory[index];
                            return ListTile(
                              leading:
                                  const Icon(Icons.history, color: Colors.teal),
                              title: Text(
                                item['user_text'] ?? 'Bilinmeyen Şikayet',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                    fontWeight: FontWeight.w500),
                              ),
                              subtitle: Text(
                                'Kategori: ${item['category'] ?? '-'} \nDurum: ${item['status'] ?? '-'}',
                              ),
                              isThreeLine: true,
                              onTap: () {},
                            );
                          },
                        ),
            ),
          ],
        ),
      ),

      body: Column(
        children: [
          // 1. MESAJLARIN AKTIĞI ALAN
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                final message = _messages[index];

                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  child: Row(
                    mainAxisAlignment: message.isUser
                        ? MainAxisAlignment.end // Bizim mesajlar sağa yaslı
                        : MainAxisAlignment.start, // Botun mesajları sola yaslı
                    crossAxisAlignment: CrossAxisAlignment
                        .end, // İkon balonun altında hizalansın
                    children: [
                      // EĞER MESAJ BOT'A AİTSE PROFİL FOTOSU KOY
                      if (!message.isUser) ...[
                        const CircleAvatar(
                          backgroundColor: Colors.teal,
                          radius: 16,
                          child: Icon(Icons.smart_toy,
                              color: Colors.white, size: 20),
                        ),
                        const SizedBox(
                            width: 8), // Fotoğraf ile balon arası boşluk
                      ],

                      // MESAJ BALONCUK KISMI
                      Flexible(
                        child: Column(
                          crossAxisAlignment: message.isUser
                              ? CrossAxisAlignment.end
                              : CrossAxisAlignment.start,
                          children: [
                            // YENİ: Eğer mesajda Base64 resim varsa göster
                            if (message.base64Image != null)
                              Container(
                                margin: const EdgeInsets.only(bottom: 4),
                                constraints: const BoxConstraints(
                                    maxHeight: 200, maxWidth: 200),
                                decoration: BoxDecoration(
                                  borderRadius: BorderRadius.circular(12),
                                  image: DecorationImage(
                                    image: MemoryImage(
                                        base64Decode(message.base64Image!)),
                                    fit: BoxFit.cover,
                                  ),
                                ),
                              ),
                            if (message.text.isNotEmpty)
                              Container(
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 16, vertical: 12),
                                decoration: BoxDecoration(
                                  color: message.isUser
                                      ? Colors.teal
                                      : Colors.white,
                                  borderRadius: BorderRadius.only(
                                    topLeft: const Radius.circular(16),
                                    topRight: const Radius.circular(16),
                                    bottomLeft: Radius.circular(
                                        message.isUser ? 16 : 0),
                                    bottomRight: Radius.circular(
                                        message.isUser ? 0 : 16),
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: Colors.black.withOpacity(0.05),
                                      blurRadius: 5,
                                      offset: const Offset(0, 2),
                                    )
                                  ],
                                ),
                                child: Text(
                                  message.text,
                                  style: TextStyle(
                                    color: message.isUser
                                        ? Colors.white
                                        : Colors.black87,
                                    fontSize: 16,
                                  ),
                                ),
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),

          // 2. YÜKLENİYOR İNDİKATÖRÜ (Bot düşünürken)
          if (_isLoading)
            const Padding(
              padding: EdgeInsets.all(8.0),
              child: Align(
                alignment: Alignment.centerLeft,
                child: CircularProgressIndicator(strokeWidth: 3),
              ),
            ),

          // YENİ: SEÇİLEN FOTOĞRAFIN GİRİŞ ALANI ÜZERİNDEKİ ÖN İZLEMESİ
          if (_selectedImage != null)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              color: Colors.grey.shade200,
              child: Row(
                children: [
                  Image.file(_selectedImage!,
                      height: 50, width: 50, fit: BoxFit.cover),
                  const SizedBox(width: 12),
                  const Text("Fotoğraf eklendi",
                      style: TextStyle(fontWeight: FontWeight.w500)),
                  const Spacer(),
                  IconButton(
                    icon: const Icon(Icons.close, color: Colors.red),
                    onPressed: () => setState(() {
                      _selectedImage = null;
                      _base64ImageString = null;
                    }),
                  )
                ],
              ),
            ),

          // 3. MESAJ YAZMA VE GÖNDERME KUTUSU
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                    color: Colors.grey.shade300,
                    blurRadius: 4,
                    offset: const Offset(0, -1))
              ],
            ),
            child: Row(
              children: [
                // YENİ: Fotoğraf Seçme Butonu (+)
                IconButton(
                  icon: const Icon(Icons.add_a_photo, color: Colors.teal),
                  onPressed: () {
                    showModalBottomSheet(
                      context: context,
                      builder: (context) => SafeArea(
                        child: Wrap(
                          children: [
                            ListTile(
                              leading: const Icon(Icons.photo_library,
                                  color: Colors.teal),
                              title: const Text('Galeriden Seç'),
                              onTap: () {
                                Navigator.pop(context);
                                _pickImage(ImageSource.gallery);
                              },
                            ),
                            ListTile(
                              leading: const Icon(Icons.camera_alt,
                                  color: Colors.teal),
                              title: const Text('Kamera ile Çek'),
                              onTap: () {
                                Navigator.pop(context);
                                _pickImage(ImageSource.camera);
                              },
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
                Expanded(
                  child: TextField(
                    controller: _controller,
                    textInputAction: TextInputAction.send,
                    onSubmitted: (_) => _sendMessage(),
                    decoration: InputDecoration(
                      hintText: 'Şikayetinizi yazın...',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(24),
                        borderSide: BorderSide.none,
                      ),
                      filled: true,
                      fillColor: Colors.grey.shade100,
                      contentPadding: const EdgeInsets.symmetric(
                          horizontal: 20, vertical: 10),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                CircleAvatar(
                  backgroundColor: Colors.teal,
                  radius: 24,
                  child: IconButton(
                    icon: const Icon(Icons.send, color: Colors.white),
                    onPressed: _sendMessage,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ============================================================================
// YENİ AYRI TASARIM: YÖNETİM PANELİ EKRANI (DASHBOARD)
// ============================================================================
class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  List<dynamic> _complaints = [];
  bool _isLoading = true;
  bool _isAuthorized = false; // <--- NOKTALI VİRGÜL EKLENDİ

  // İstatistik sayaçları
  int totalComplaints = 0;
  int highPriorityCount = 0;
  Map<String, int> departmentStats = {};

  @override
  void initState() {
    super.initState();
    _loadDashboardData();
    // EKRAN AÇILDIĞI AN KİLİDİ TETİKLE:
    WidgetsBinding.instance.addPostFrameCallback((_) => _showPasswordDialog(context));
  }

  // KİLİT FONKSİYONU
  void _showPasswordDialog(BuildContext context) {
    final TextEditingController _passController = TextEditingController();
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        title: const Text("Yönetici Girişi"),
        content: TextField(controller: _passController, obscureText: true, decoration: const InputDecoration(labelText: "Şifre")),
        actions: [
          TextButton(
            onPressed: () async {
              final url = Uri.parse('https://citypulse-backend-wjt6.onrender.com/api/login?password=${_passController.text}');
              try {
                final response = await http.get(url);
                if (response.statusCode == 200) {
                  setState(() => _isAuthorized = true);
                  Navigator.of(context).pop();
                } else {
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("Hatalı Şifre!")));
                }
              } catch (e) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("Bağlantı hatası!")));
              }
            },
            child: const Text("Giriş Yap"),
          ),
        ],
      ),
    );
  }

  Future<void> _loadDashboardData() async {
    try {
      final url = Uri.parse('https://citypulse-backend-wjt6.onrender.com/api/complaints');
      final response = await http.get(url);
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as List;
        int high = 0;
        Map<String, int> depts = {};
        for (var item in data) {
          if (item['priority'].toString().contains('Yüksek')) high++;
          String dept = item['department'] ?? 'Belirsiz';
          depts[dept] = (depts[dept] ?? 0) + 1;
        }
        setState(() {
          _complaints = data;
          totalComplaints = data.length;
          highPriorityCount = high;
          departmentStats = depts;
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      appBar: AppBar(title: const Text('Belediye Yönetim Paneli'), backgroundColor: Colors.teal.shade100),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : !_isAuthorized 
              ? _buildLockedScreen() 
              : _buildDashboardContent(),
    );
  }

  Widget _buildLockedScreen() => const Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [Icon(Icons.lock, size: 80, color: Colors.teal), Text("Erişim için şifre girin")]));

  Widget _buildDashboardContent() {
    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: ListView(
        children: [
          const Text('Şehir Genel Durum Özetleri',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.black87)),
          const SizedBox(height: 16),
          
          // Rapor Butonu
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4.0),
            child: SizedBox(
              width: double.infinity,
              height: 50,
              child: ElevatedButton.icon(
                onPressed: () async {
                  final Uri url = Uri.parse('https://citypulse-backend-wjt6.onrender.com/api/export');
                  if (await canLaunchUrl(url)) {
                    await launchUrl(url, mode: LaunchMode.externalApplication);
                  }
                },
                icon: const Icon(Icons.download_rounded, color: Colors.white),
                label: const Text("Tüm Şikayetleri Excel Olarak İndir",
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF2A4B7C),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
              ),
            ),
          ),
          const SizedBox(height: 24),

          // 1. İSTATİSTİK KARTLARI
          Row(
            children: [
              Expanded(
                child: _buildStatCard('Toplam Şikayet', totalComplaints.toString(), Colors.blue.shade700, Icons.assignment),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildStatCard('Acil (Yüksek)', highPriorityCount.toString(), Colors.red.shade700, Icons.warning_amber_rounded),
              ),
            ],
          ),
          const SizedBox(height: 24),

          // 2. BİRİMLERE GÖRE DAĞILIM PANELİ
          const Text('Birimlere Göre Şikayet Yükü', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.black87)),
          const SizedBox(height: 8),
          Card(
            color: Colors.white,
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: departmentStats.isEmpty
                  ? const Text('Henüz veri yok.')
                  : Column(
                      children: departmentStats.entries.map((entry) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 6.0),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(entry.key, style: const TextStyle(fontWeight: FontWeight.w500)),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                decoration: BoxDecoration(color: Colors.teal.shade50, borderRadius: BorderRadius.circular(12)),
                                child: Text('${entry.value} Adet', style: const TextStyle(color: Colors.teal, fontWeight: FontWeight.bold)),
                              )
                            ],
                          ),
                        );
                      }).toList(),
                    ),
            ),
          ),
          const SizedBox(height: 24),

          // 3. EN SON GELEN ŞİKAYETLERİN LİSTESİ
          const Text('Gelen Son Şikayetler (Log Akışı)', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.black87)),
          const SizedBox(height: 8),
          ..._complaints.take(5).map((item) {
            return Card(
              color: Colors.white,
              margin: const EdgeInsets.symmetric(vertical: 4),
              child: ListTile(
                leading: const CircleAvatar(backgroundColor: Colors.teal, child: Icon(Icons.mail_outline, color: Colors.white)),
                title: Text(item['user_text'] ?? '-', maxLines: 1, overflow: TextOverflow.ellipsis),
                subtitle: Text('Birim: ${item['department']} \nÖncelik: ${item['priority']}'),
                trailing: Text(item['status'] ?? 'Beklemede', style: const TextStyle(color: Colors.orange, fontWeight: FontWeight.bold)),
              ),
            );
          }).toList(),
        ],
      ),
    );
  }

  Widget _buildStatCard(
      String title, String value, Color color, IconData icon) {
    return Card(
      color: Colors.white,
      elevation: 2,
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: color, size: 32),
            const SizedBox(height: 12),
            Text(title,
                style: TextStyle(color: Colors.grey.shade600, fontSize: 14)),
            const SizedBox(height: 4),
            Text(value,
                style: const TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.bold,
                    color: Colors.black87)),
          ],
        ),
      ),
    );
  }
}
