package src

import (
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

type EarthQuake struct {
	EventID        string  `json:"eventID"`
	RMS            string  `json:"rms"`
	DateTime       string  `json:"date"`
	Location       string  `json:"location"`
	Magnitude      string  `json:"magnitude"`
	Latitude       string  `json:"latitude"`
	Longitude      string  `json:"longitude"`
	Depth          string  `json:"depth"`
	Country        string  `json:"country"`
	Province       string  `json:"province"`
	District       string  `json:"district"`
	Neighborhood   string  `json:"neighborhood"`
	IsEventUpdate  bool    `json:"isEventUpdate"`
	LastUpdateDate *string `json:"lastUpdateDate"`
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

	// log.Println("API Response Body:")
	// log.Println(string(bodyBytes))

	log.Println("Application working...")

	var earthquakes []EarthQuake
	if err := json.Unmarshal(bodyBytes, &earthquakes); err != nil {
		http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	if len(earthquakes) == 0 {
		fmt.Fprint(w, "<tr><td colspan='10'>No data found.</td></tr>")
		return
	}

	for _, eq := range earthquakes {
		date := ""
		timeStr := ""
		if len(eq.DateTime) >= 19 {
			date = eq.DateTime[:10]
			timeStr = eq.DateTime[11:19]
		} else {
			date = eq.DateTime
		}

		fmt.Fprintf(w, "<tr>"+
			"<td>%s</td>"+ // Date
			"<td>%s</td>"+ // Time
			"<td>%s</td>"+ // Latitude
			"<td>%s</td>"+ // Longitude
			"<td>%s</td>"+ // Depth
			"<td>%s</td>"+ // Magnitude
			"<td>%s</td>"+ // Location
			"<td>%s</td>"+ // Country/Province
			"<td>%s</td>"+ // District
			"<td>%s</td>"+ // Neighborhood
			"</tr>\n",
			date, timeStr, eq.Latitude, eq.Longitude, eq.Depth, eq.Magnitude, eq.Location,
			eq.Province, eq.District, eq.Neighborhood)
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
