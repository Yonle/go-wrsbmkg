package wrsbmkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Untuk beberapa alasan, Tipe data json akan seperti ini.
// Pemakaian setiap data dapat dilihat dari JSON yang direturn oleh setiap endpoint.
//
// Data ini bisa diparse dengan [codeberg.org/Yonle/go-wrsbmkg/helper].
type DataJSON map[string]interface{}

// Penerima data gempa yang akan diambil dari API BMKG.
//
// Pastikan bahwa API_URL dan Interval sudah disertakan.
// Interval yang disarankan adalah `time.Second*15`
type Penerima struct {
	// Interval penerimaan informasi baru
	Interval time.Duration

	GempaTerakhir    DataJSON
	RealtimeTerakhir DataJSON
	NarasiTerakhir   string

	Gempa    chan DataJSON
	Realtime chan DataJSON
	Narasi   chan string

	API_URL string

	// Timeout dan segala lainnya berkaitan request, Atur dengan http.Client
	HTTP_Client http.Client
}

var DEFAULT_API_URL string = "https://bmkg-content-inatews.storage.googleapis.com"

// Unix Milli Now
func umn() int64 {
	now := time.Now()
	return now.UnixMilli()
}

func (p *Penerima) PollingGempa(ctx context.Context) {
	var identifierTerakhir string

listener:
	for {
		select {
		case <-ctx.Done():
			break listener
		case <-time.After(p.Interval):
			j, _, err := p.DownloadGempa(ctx)
			if err != nil {
				continue listener
			}

			i := j["identifier"].(string)

			if identifierTerakhir == i {
				continue listener
			}

			identifierTerakhir = i
			p.Gempa <- j
			p.GempaTerakhir = j

			continue listener
		}
	}
}

func (p *Penerima) PollingRealtime(ctx context.Context) {
	var informasiTerakhir string

listener:
	for {
		select {
		case <-ctx.Done():
			break listener
		case <-time.After(p.Interval):
			j, _, err := p.DownloadRealtime(ctx)
			if err != nil {
				continue listener
			}

			r := j["features"].([]interface{})

			a := r[0].(map[string]interface{})

			d := a["properties"].(map[string]any)
			t := d["time"].(string)

			if informasiTerakhir == t {
				continue listener
			}

			informasiTerakhir = t
			p.Realtime <- j
			p.RealtimeTerakhir = j
			continue listener
		}
	}
}

func (p *Penerima) PollingNarasi(ctx context.Context) {
	var gempaTerakhir int64

listener:
	for {
		select {
		case <-ctx.Done():
			break listener
		case <-time.After(p.Interval):
			if p.GempaTerakhir == nil {
				continue listener
			}

			i := p.GempaTerakhir["info"].(map[string]interface{})
			d := i["eventid"].(string)
			t, err := strconv.ParseInt(d, 10, 64)

			if err != nil {
				panic(err)
			}

			if gempaTerakhir == t {
				continue listener
			}

			narasi, resp, err := p.DownloadNarasi(ctx, t)
			if err != nil {
				continue listener
			}

			if resp.StatusCode != 200 {
				continue listener
			}

			gempaTerakhir = t
			p.Narasi <- narasi
			p.NarasiTerakhir = narasi
			continue listener
		}
	}
}

// Fungsi ini akan mulai menerima data baru setiap waktu.
// Sebelum memanggil, Pastikan bahwa Interval sudah ditentukan di Penerima{}.
//
// Hasil dari channel [DataJSON] akan sama dengan apa yang dihasilkan oleh fungsi [DownloadGempa] dan [DownloadRealtime].
//
// Jangan panggil fungsi ini jika Penerima sudah dijalankan, Kecuali sudah dihentikan dengan [context.Context].
// Disarankan untuk menggunakan [context.WithCancel] untuk menghentikan penerimaan data.
func (p *Penerima) MulaiPolling(ctx context.Context) error {
	go p.PollingGempa(ctx)
	go p.PollingRealtime(ctx)
	go p.PollingNarasi(ctx)
	return nil
}

func (p *Penerima) Get(ctx context.Context, path string) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.API_URL+path, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, getErr := p.HTTP_Client.Do(req)
	if getErr != nil {
		return nil, nil, getErr
	}

	b, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()

	if readErr != nil {
		return nil, resp, readErr
	}

	return b, resp, nil
}

func (p *Penerima) GetJSON(ctx context.Context, path string) (DataJSON, *http.Response, error) {
	b, resp, err := p.Get(ctx, path)

	if err != nil {
		return nil, resp, err
	}

	var j DataJSON
	parseErr := json.Unmarshal(b, &j)
	if parseErr != nil {
		return nil, resp, parseErr
	}

	return j, resp, nil
}

// Ini akan mendownload informasi gempa.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/datagempa.json
func (p *Penerima) DownloadGempa(ctx context.Context) (DataJSON, *http.Response, error) {
	n := umn()
	path := fmt.Sprintf("/datagempa.json?t=%d", n)
	return p.GetJSON(ctx, path)
}

// Ini akan mendownload data gempa realtime.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/lastQL.json
func (p *Penerima) DownloadRealtime(ctx context.Context) (DataJSON, *http.Response, error) {
	n := umn()
	path := fmt.Sprintf("/lastQL.json?t=%d", n)
	return p.GetJSON(ctx, path)
}

// Ini akan mendownload riwayat data gempa.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/gempaQL.json
func (p *Penerima) DownloadRiwayatGempa(ctx context.Context) (DataJSON, *http.Response, error) {
	return p.GetJSON(ctx, "/gempaQL.json")
}

// Ini akan mendownload teks narasi.
// Setiap narasi tidak langsung tersedia setelah peringatan gempa diumumkan, Melainkan memerlukan beberapa waktu.
//
// Teks narasi yang diterima berbentuk HTML.
// Elemen HTML dapat dihilangkan dengan memakai [codeberg.org/Yonle/go-wrsbmkg/helper].
func (p *Penerima) DownloadNarasi(ctx context.Context, eventid int64) (narasi string, resp *http.Response, err error) {
	path := fmt.Sprintf("/%d_narasi.txt", eventid)
	b, resp, err := p.Get(ctx, path)
	if err != nil {
		return "", resp, err
	}

	teks := string(b)
	return teks, resp, nil
}
