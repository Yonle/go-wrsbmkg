package wrsbmkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ERR_TIMEOUT_WAITING_NARASI = errors.New("Timed out waiting for narasi")

// Penerima data gempa yang akan diambil dari API BMKG.
//
// Pastikan bahwa API_URL dan Interval sudah disertakan.
// Interval yang disarankan adalah `time.Second*15`
type Penerima struct {
	// Interval penerimaan informasi baru
	Interval time.Duration

	// Struct Raw_* dapat disimplikasi ke struct yang mudah dipakai
	// Dengan memakai modul codeberg.org/Yonle/go-wrsbmkg/helper
	GempaTerakhir    *Raw_DataGempa
	RealtimeTerakhir *Raw_QL

	Gempa    chan *Raw_DataGempa
	Realtime chan *Raw_QL

	API_URL string

	// Timeout dan segala lainnya berkaitan request, Atur dengan http.Client
	HTTP_Client *http.Client
}

// Ini akan memuat [Penerima] dengan parameter default.
func BuatPenerima() *Penerima {
	return &Penerima{
		Gempa:    make(chan *Raw_DataGempa),
		Realtime: make(chan *Raw_QL),

		Interval: time.Second * 15,
		API_URL:  DEFAULT_API_URL,

		HTTP_Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
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

			i := j.Identifier

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

			r := j.Features

			a := r[0]

			d := a.Properties
			t := d.Time

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

// FetchNarasi() untuk menunggu & menerima pesan narasi dari pihak BMKG setelah menerima alert.
// Fungsi ini bersifat synchronous. Disarankan untuk dijalankan dengan goroutine.
// Jika sudah melewati [deadline], Maka [err] akan mengreturn [ERR_TIMEOUT_WAITING_NARASI]
func (p *Penerima) FetchNarasi(c context.Context, eventid string, deadline time.Time) (string, error) {
	ctx, cancel := context.WithDeadlineCause(c, deadline, ERR_TIMEOUT_WAITING_NARASI)
	defer cancel()

listener:
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(p.Interval):
			narasi, resp, err := p.DownloadNarasi(ctx, eventid)
			if err != nil {
				continue listener
			}

			if resp.StatusCode != 200 {
				continue listener
			}

			return narasi, nil
		}
	}
}

// Fungsi ini akan mulai menerima data baru setiap waktu.
// Sebelum memanggil, Pastikan bahwa Interval sudah ditentukan di Penerima{}.
//
// Jangan panggil fungsi ini jika Penerima sudah dijalankan, Kecuali sudah dihentikan dengan [context.Context].
// Disarankan untuk menggunakan [context.WithCancel] untuk menghentikan penerimaan data.
func (p *Penerima) MulaiPolling(ctx context.Context) error {
	go p.PollingGempa(ctx)
	go p.PollingRealtime(ctx)
	return nil
}

func (p *Penerima) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.API_URL+path, nil)
	if err != nil {
		return nil, err
	}

	return p.HTTP_Client.Do(req)
}

func (p *Penerima) GetBody(ctx context.Context, path string) ([]byte, *http.Response, error) {
	resp, err := p.Get(ctx, path)
	if err != nil {
		return nil, resp, err
	}

	defer resp.Body.Close()

	b, readErr := io.ReadAll(resp.Body)
	return b, resp, readErr
}

// Ini akan mendownload informasi gempa.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/datagempa.json
func (p *Penerima) DownloadGempa(ctx context.Context) (*Raw_DataGempa, *http.Response, error) {
	n := umn()
	path := fmt.Sprintf("/datagempa.json?t=%d", n)
	resp, err := p.Get(ctx, path)
	if err != nil {
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, resp, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	var dg Raw_DataGempa

	if err := json.NewDecoder(resp.Body).Decode(&dg); err != nil {
		return nil, resp, err
	}

	return &dg, resp, nil
}

// Ini akan mendownload data gempa realtime.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/lastQL.json
func (p *Penerima) DownloadRealtime(ctx context.Context) (*Raw_QL, *http.Response, error) {
	n := umn()
	path := fmt.Sprintf("/lastQL.json?t=%d", n)
	return p.downloadQL(ctx, path)
}

// Ini akan mendownload riwayat data gempa.
// Lihat data JSON asli di https://bmkg-content-inatews.storage.googleapis.com/gempaQL.json
func (p *Penerima) DownloadRiwayatGempa(ctx context.Context) (*Raw_QL, *http.Response, error) {
	return p.downloadQL(ctx, "/gempaQL.json")
}

func (p *Penerima) downloadQL(ctx context.Context, path string) (*Raw_QL, *http.Response, error) {
	resp, err := p.Get(ctx, path)
	if err != nil {
		return nil, resp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, resp, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}

	var q Raw_QL
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return nil, resp, err
	}

	return &q, resp, nil
}

// Ini akan mendownload teks narasi.
// Setiap narasi tidak langsung tersedia setelah peringatan gempa diumumkan, Melainkan memerlukan beberapa waktu.
//
// Teks narasi yang diterima berbentuk HTML.
// Elemen HTML dapat dihilangkan dengan memakai [codeberg.org/Yonle/go-wrsbmkg/helper].
func (p *Penerima) DownloadNarasi(ctx context.Context, eventid string) (narasi string, resp *http.Response, err error) {
	path := fmt.Sprintf("/%s_narasi.txt", eventid)
	b, resp, err := p.GetBody(ctx, path)
	if err != nil {
		return "", resp, err
	}

	teks := string(b)
	return teks, resp, nil
}
