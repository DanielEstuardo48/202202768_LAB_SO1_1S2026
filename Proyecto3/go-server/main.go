package main

import(
	"fmt"
	"context"
	"encoding/json"
	"log"
	"net"

	pb "go-server/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// Estructura del Json para RabbitMQ
type WarReportMessage struct {
	Country          string `json:"country"`
	WarplanesInAir   int32  `json:"warplanes_in_air"`
	WarshipsInWater  int32  `json:"warships_in_water"`
	Timestamp        string `json:"timestamp"`
}

type server struct {
	pb.UnimplementedWarReportServiceServer
	rabbitChannel *amqp.Channel
}

func (s *server) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	log.Printf("Reporte recibido via gRPC: %v", req)

	// Convertir el enum de país a string
	countryName := req.Country.String()

	// Crear el mensaje para RabbitMQ
	msg := WarReportMessage{
		Country:         countryName,
		WarplanesInAir:  req.WarplanesInAir,
		WarshipsInWater: req.WarshipsInWater,
		Timestamp:       req.Timestamp,
	}

	// Serializar a JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error serializando mensaje: %v", err)
	}

	// Publicar en RabbitMQ
	err = s.rabbitChannel.Publish(
		"",             
		"war_reports",  // nombre de la cola
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error publicando en RabbitMQ: %v", err)
	}

	log.Printf("Mensaje publicado en RabbitMQ: %s", body)

	return &pb.WarReportResponse{Status: "ok"}, nil
}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	
	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}

	// Declarar la cola — si ya existe no hace nada
	_, err = ch.QueueDeclare(
		"war_reports", // nombre
		true,          // durable (sobrevive reinicios)
		false,         // auto-delete
		false,         // exclusive
		false,         // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Error declarando cola: %v", err)
	}

	log.Println("Conectado a RabbitMQ correctamente")
	return conn, ch
}

func main() {
	// Conectar a RabbitMQ
	_, ch := connectRabbitMQ()

	// Iniciar servidor gRPC en puerto 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error iniciando listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWarReportServiceServer(grpcServer, &server{rabbitChannel: ch})

	log.Println("gRPC Server corriendo en puerto 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error en servidor gRPC: %v", err)
	}
}