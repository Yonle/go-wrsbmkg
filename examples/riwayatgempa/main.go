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
		API_URL: wrsbmkg.DEFAULT_API_URL,

		HTTP_Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	ctx := context.Background()
	raw_riwayat, _, err := p.DownloadRiwayatGempa(ctx)
	if err != nil {
		panic(err)
	}

	riwayat := helper.ParseRiwayatGempa(raw_riwayat)

	for _, realtime := range riwayat {
		fmt.Println("---")
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
	}
}
