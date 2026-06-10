package repository

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global değişkenler, bu sayede servisin içinden bunlara erişebileceğiz
var MongoClient *mongo.Client
var ComplaintCollection *mongo.Collection

// MongoDB Atlas'a bağlanan fonksiyon
func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("🔴 HATA: .env dosyasında MONGO_URI bulunamadı!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("🔴 Veritabanına bağlanılamadı: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("🔴 MongoDB Atlas'a ping atılamadı: %v", err)
	}

	log.Println("🟢 MongoDB Atlas Bağlantısı Başarılı!")

	MongoClient = client
	ComplaintCollection = client.Database("citypulse").Collection("complaints")
}