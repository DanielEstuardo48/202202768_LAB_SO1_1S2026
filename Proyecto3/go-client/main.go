package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "go-client/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Estructura para decodificar el JSON que viene de Rust
type WarReport struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

var grpcClient pb.WarReportServiceClient

func receiveFromRust(w http.ResponseWriter, r *http.Request) {
	// Decodificar el JSON que viene de Rust
	var report WarReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	log.Printf("Recibido de Rust: %+v", report)

	// Convertir el string de país al enum del proto
	countryEnum := stringToCountry(report.Country)

	// Llamar al gRPC Server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcClient.SendReport(ctx, &pb.WarReportRequest{
		Country:         countryEnum,
		WarplanesInAir:  report.WarplanesInAir,
		WarshipsInWater: report.WarshipsInWater,
		Timestamp:       report.Timestamp,
	})

	if err != nil {
		log.Printf("Error llamando gRPC: %v", err)
		http.Error(w, "Error procesando reporte", http.StatusInternalServerError)
		return
	}

	log.Printf("Respuesta gRPC: %s", resp.Status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": resp.Status})
}

// Convierte el string del país al enum generado por protoc
func stringToCountry(country string) pb.Country {
	switch country {
	case "USA":
		return pb.Country_USA
	case "RUS":
		return pb.Country_RUS
	case "CHN":
		return pb.Country_CHN
	case "ESP":
		return pb.Country_ESP
	case "GTM":
		return pb.Country_GTM
	default:
		return pb.Country_UNKNOWN
	}
}

func main() {
	// Conectar al gRPC Server
	conn, err := grpc.Dial(
		"go-server:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Error conectando al gRPC server: %v", err)
	}
	defer conn.Close()

	grpcClient = pb.NewWarReportServiceClient(conn)
	log.Println("Conectado al gRPC Server")

	// Servidor HTTP que escucha peticiones de Rust
	http.HandleFunc("/report", receiveFromRust)
	log.Println("Go Client corriendo en puerto 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}