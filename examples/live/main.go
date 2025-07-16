package main

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/Yonle/go-wrsbmkg"
	"codeberg.org/Yonle/go-wrsbmkg/helper"
)

var narasi = make(chan string)

func main() {
	p := wrsbmkg.BuatPenerima()

	ctx := context.Background()
	p.MulaiPolling(ctx)

	fmt.Println("WRS-BMKG")
	fmt.Println("Informasi akan dimuat dalam 15 detik....")

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

			go func() {
				teksNarasi, err := p.FetchNarasi(ctx, g.Info.EventID, time.Now().Add(time.Hour))
				if err != nil {
					return
				}

				narasi <- teksNarasi
			}()
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
		case n := <-narasi:
			fmt.Println("\nNARASI ---")

			narasi := helper.CleanNarasi(n)
			fmt.Println(narasi)
		}
	}
}
