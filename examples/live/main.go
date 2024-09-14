package main

import (
	"codeberg.org/Yonle/go-wrsbmkg"
	"codeberg.org/Yonle/go-wrsbmkg/helper"
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	p := wrsbmkg.Penerima{
		Gempa:    make(chan wrsbmkg.DataJSON),
		Realtime: make(chan wrsbmkg.DataJSON),
		Narasi:   make(chan string),

		Interval: time.Second * 15,
		API_URL:  wrsbmkg.DEFAULT_API_URL,

		HTTP_Client: http.Client{
			Timeout: time.Second * 30,
		},
	}

	ctx := context.Background()
	p.MulaiPolling(ctx)

	fmt.Println("WRS-BMKG")

	for {
		fmt.Println("---")
		select {
		case g := <-p.Gempa:
			gempa := helper.ParseGempa(g)

			fmt.Println("\nGEMPABUMI ---")
			fmt.Printf(
				"%s\n\n%s\n\n%s\n\n%s\n\n%s\n",
				gempa.Subject,
				gempa.Description,
				gempa.Area,
				gempa.Potential,
				gempa.Instruction,
			)
		case r := <-p.Realtime:
			realtime := helper.ParseRealtime(r)
			fmt.Println("\nREALTIME ---")

			fmt.Printf(
				"%s\n"+
					"Tanggal   : %s\n"+
					"Magnitudo : %v\n"+
					"Kedalaman : %v\n"+
					"Koordinat : %s,%s\n"+
					"Fase      : %v\n"+
					"Status    : %s\n",
				realtime.Place,
				realtime.Time,
				realtime.Magnitude,
				realtime.Depth,
				realtime.Coordinates[1].(string),
				realtime.Coordinates[0].(string),
				realtime.Phase,
				realtime.Status,
			)
		case n := <-p.Narasi:
			fmt.Println("\nNARASI ---")

			narasi := helper.CleanNarasi(n)
			fmt.Println(narasi)
		}
	}
}
