package handlers

import (
	"net/http"
	"os"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	adminPass := os.Getenv("ADMIN_PASSWORD")

	// Eğer env boşsa, girişi direkt reddet!
	// Böylece kodda şifre olması imkansız hale gelir.
	if adminPass == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Sunucu hatasi: Admin sifresi tanimlanmamis."))
		return
	}

	password := r.URL.Query().Get("password")

	if password == adminPass {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Giris basarili"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Yanlis sifre"))
	}
}
