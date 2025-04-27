package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	baseAPI   = "https://deprem.afad.gov.tr/apiv2/event/filter"
	port      = "8082"
	localHost = "http://localhost:" + port
)

type AfadResponse struct {
	Data []earthQuake `json:"data"`
}

type earthQuake struct {
	EventID   int     `json:"EventID"`
	DateTime  string  `json:"DateTime"`
	Location  string  `json:"Location"`
	Magnitude float64 `json:"Magnitude"`
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
	Depth     float64 `json:"Depth"`
}

func earthQuakeHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	start := query.Get("start")
	end := query.Get("end")

	if start == "" || end == "" {
		now := time.Now().UTC()
		start = now.Add(-24 * time.Hour).Format("2006-01-02T15:04:05")
		end = now.Format("2006-01-02T15:04:05")
	}

	apiURL, err := url.Parse(baseAPI)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	params.Set("limit", "100")
	params.Set("orderby", "timedesc")

	apiURL.RawQuery = params.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		http.Error(w, "Failed to fetch data.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API returned error.", http.StatusBadGateway)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body.", http.StatusInternalServerError)
		return
	}

	log.Println("API Response Body:")
	log.Println(string(bodyBytes))

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var afadData AfadResponse
	if err := json.NewDecoder(resp.Body).Decode(&afadData); err != nil {
		http.Error(w, "Failed to parse JSON.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	if len(afadData.Data) == 0 {
		fmt.Fprint(w, "<tr><td colspan='8'>No data found.</td></tr>")
		return
	}

	for _, d := range afadData.Data {
		date := ""
		timeStr := ""
		if len(d.DateTime) >= 19 {
			date = d.DateTime[:10]
			timeStr = d.DateTime[11:19]
		} else {
			date = d.DateTime
		}

		fmt.Fprintf(w, "<tr>"+
			"<td>%s</td>"+
			"<td>%s</td>"+
			"<td>%.4f</td>"+
			"<td>%.4f</td>"+
			"<td>%.1f</td>"+
			"<td>%.1f</td>"+
			"<td>%s</td>"+
			"<td>-</td>"+
			"</tr>\n",
			date, timeStr, d.Latitude, d.Longitude, d.Depth, d.Magnitude, d.Location)
	}
}

func RunRequest() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	go openBrowser(localHost + "/static")

	http.HandleFunc("/eq-rows", earthQuakeHandler)
	log.Printf("Server run: %v\n", localHost)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
