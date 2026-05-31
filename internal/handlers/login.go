package handlers

import (
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	// Burada admin şifresini kontrol et
	if password == "123456" { // Ya da os.Getenv("ADMIN_PASSWORD")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Giris basarili"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Yanlis sifre"))
	}
}