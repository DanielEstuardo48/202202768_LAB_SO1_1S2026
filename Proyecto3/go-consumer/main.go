package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valkey-io/valkey-go"
)

type WarReportMessage struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Error conectando RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}
	_, err = ch.QueueDeclare("war_reports", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error declarando cola: %v", err)
	}
	return conn, ch
}

func connectValkey() valkey.Client {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		addr = "valkey:6379"
	}
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		log.Fatalf("Error conectando Valkey: %v", err)
	}
	log.Printf("Conectado a Valkey en %s", addr)
	return client
}

func recalcStats(ctx context.Context, client valkey.Client, listKey, maxKey, minKey, modaKey string) {
	result, err := client.Do(ctx, client.B().Lrange().Key(listKey).Start(0).Stop(-1).Build()).AsStrSlice()
	if err != nil || len(result) == 0 {
		return
	}

	freq := make(map[int]int)
	maxVal := -1
	minVal := 999999

	for _, s := range result {
		v, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		freq[v]++
		if v > maxVal {
			maxVal = v
		}
		if v < minVal {
			minVal = v
		}
	}

	moda := 0
	maxFreq := 0
	for val, f := range freq {
		if f > maxFreq {
			maxFreq = f
			moda = val
		}
	}

	client.Do(ctx, client.B().Set().Key(maxKey).Value(fmt.Sprintf("%d", maxVal)).Build())
	client.Do(ctx, client.B().Set().Key(minKey).Value(fmt.Sprintf("%d", minVal)).Build())
	client.Do(ctx, client.B().Set().Key(modaKey).Value(fmt.Sprintf("%d", moda)).Build())
}

func saveToValkey(client valkey.Client, report WarReportMessage) {
	ctx := context.Background()

	// 1. Contador por país
	client.Do(ctx, client.B().Incr().Key("reports:count:"+report.Country).Build())

	// 2. Último reporte del país
	data, _ := json.Marshal(report)
	client.Do(ctx, client.B().Set().Key("reports:last:"+report.Country).Value(string(data)).Build())

	// 3. Timeline listas separadas para aviones y barcos
	planesTimelineKey := "timeline:planes:" + report.Country
	shipsTimelineKey := "timeline:ships:" + report.Country

	client.Do(ctx, client.B().Lpush().Key(planesTimelineKey).
		Element(fmt.Sprintf("%d", report.WarplanesInAir)).Build())
	client.Do(ctx, client.B().Ltrim().Key(planesTimelineKey).Start(0).Stop(99).Build())

	client.Do(ctx, client.B().Lpush().Key(shipsTimelineKey).
		Element(fmt.Sprintf("%d", report.WarshipsInWater)).Build())
	client.Do(ctx, client.B().Ltrim().Key(shipsTimelineKey).Start(0).Stop(99).Build())

	// 4. Sorted sets con timestamp para serie temporal en Grafana
	//ts := float64(time.Now().UnixMilli())
	entry := fmt.Sprintf("%d:%d", time.Now().UnixMilli(), report.WarplanesInAir)
	client.Do(ctx, client.B().Lpush().Key("ts:planes:"+report.Country).Element(entry).Build())
	client.Do(ctx, client.B().Ltrim().Key("ts:planes:"+report.Country).Start(0).Stop(199).Build())

	entry = fmt.Sprintf("%d:%d", time.Now().UnixMilli(), report.WarshipsInWater)
	client.Do(ctx, client.B().Lpush().Key("ts:ships:"+report.Country).Element(entry).Build())
	client.Do(ctx, client.B().Ltrim().Key("ts:ships:"+report.Country).Start(0).Stop(199).Build())
	// Mantener solo los ultimos 200 registros
	client.Do(ctx, client.B().Zremrangebyrank().Key("ts:planes:"+report.Country).Start(0).Stop(-201).Build())
	client.Do(ctx, client.B().Zremrangebyrank().Key("ts:ships:"+report.Country).Start(0).Stop(-201).Build())

	// 5. Listas globales para estadísticas
	client.Do(ctx, client.B().Rpush().Key("stats:warplanes").
		Element(fmt.Sprintf("%d", report.WarplanesInAir)).Build())
	client.Do(ctx, client.B().Rpush().Key("stats:warships").
		Element(fmt.Sprintf("%d", report.WarshipsInWater)).Build())

	// 6. Sorted sets para top países
	client.Do(ctx, client.B().Zincrby().Key("top:warplanes").
		Increment(float64(report.WarplanesInAir)).Member(report.Country).Build())
	client.Do(ctx, client.B().Zincrby().Key("top:warships").
		Increment(float64(report.WarshipsInWater)).Member(report.Country).Build())

	// 7. Recalcular max, min, moda
	recalcStats(ctx, client,
		"stats:warplanes",
		"stat:planes:max",
		"stat:planes:min",
		"stat:planes:moda",
	)
	recalcStats(ctx, client,
		"stats:warships",
		"stat:ships:max",
		"stat:ships:min",
		"stat:ships:moda",
	)

	// 8. País asignado
	client.Do(ctx, client.B().Set().Key("assigned:country").Value("GTM").Build())

	log.Printf("Guardado: país=%s aviones=%d barcos=%d",
		report.Country, report.WarplanesInAir, report.WarshipsInWater)
}

func main() {
	_, ch := connectRabbitMQ()
	valkeyClient := connectValkey()
	defer valkeyClient.Close()

	msgs, err := ch.Consume("war_reports", "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error registrando consumer: %v", err)
	}

	log.Println("Consumer escuchando mensajes de RabbitMQ...")

	for msg := range msgs {
		go func(m amqp.Delivery) {
			var report WarReportMessage
			if err := json.Unmarshal(m.Body, &report); err != nil {
				log.Printf("Error deserializando: %v", err)
				return
			}
			saveToValkey(valkeyClient, report)
		}(msg)
	}
}
