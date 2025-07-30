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
				teksNarasi, err := p.FetchNarasi(ctx, g.Info.EventID, time.Now().Add(48*time.Hour))
				if err != nil {
					return
				}

				narasi <- teksNarasi
			}()

			var zonaObservasiText string

			for _, area := range gempa.ObsAreas {
				zonaObservasiText += fmt.Sprintf(
					"- %s (%s %s) dengan ketinggian %s Meter pada tanggal %s pukul %s\n",
					area.Location, area.Latitude, area.Longitude, area.Height, area.Date, area.Time,
				)
			}

			if len(zonaObservasiText) > 0 {
				headerText := "TELAH TERJADI GEMPABUMI BERPOTENSI TSUNAMI\n\n"
				headerText += fmt.Sprintf(
					"Gempa terjadi pada %s, Pukul %s, berkekuatan M%.1f, dengan kedalaman %s pada jarak %s\n",
					gempa.Date, gempa.Time, gempa.Magnitude, gempa.Depth, gempa.Area,
				)

				headerText += "\nBerdasarkan pengamatan muka air laut, tsunami telah terdeteksi di wilayah berikut:"

				zonaObservasiText = headerText + "\n" + zonaObservasiText

				fmt.Println("\n---\n" + zonaObservasiText)
			}

			var zonaPeringatanText string

			for _, area := range gempa.WZAreas {
				zonaPeringatanText += fmt.Sprintf(
					"- %s: %s, %s (estimasi waktu tiba: %s %s)\n",
					area.Level,
					area.Province,
					area.District,
					area.Date,
					area.Time,
				)
			}

			if len(zonaPeringatanText) > 0 {
				zonaPeringatanText += fmt.Sprintf(
					"\nSaran dan Arahan Status Peringatan\n1. %s\n2. %s\n3. %s",
					gempa.Instruction1,
					gempa.Instruction2,
					gempa.Instruction3,
				)

				zonaPeringatanText = "Daerah yang berpotensi tsunami berdasarkan pemodelan:\n" + zonaPeringatanText

				fmt.Println(zonaPeringatanText)
			}

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
