package main

import (
    "backendGo/internal/agent" // Modül adınla başladığından emin ol
    "fmt"
    "log"

    "github.com/joho/godotenv"
)

func main() {
	// .env dosyasını oku (Dosya bir üst dizinde olduğu için ../.env)
	godotenv.Load(".env")

	fmt.Println("CityPulse Agent Test Sistemi")
	fmt.Println("----------------------------")

	sikayet := "Yollarda çok fazla çukur var, arabamın lastiği patladı."
	
	sonuç, err := agent.AnalyzeComplaint(sikayet)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sonuç)
}