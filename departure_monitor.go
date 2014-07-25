package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
)

type DateTimeType struct {
	Date struct {
		Day   int `xml:"day,attr"`
		Month int `xml:"month,attr"`
		Year  int `xml:"year,attr"`
	} `xml:"itdDate"`

	Time struct {
		Hour   int `xml:"hour,attr"`
		Minute int `xml:"minute,attr"`
	} `xml:"itdTime"`
}

type Line struct {
	Number    string `xml:"number,attr"`
	Direction string `xml:"direction,attr"`
}

type Departure struct {
	Countdown int    `xml:"countdown,attr"`
	Platform  string `xml:"platform,attr"`

	DateTime    DateTimeType `xml:"itdDateTime"`
	ServingLine Line         `xml:"itdServingLine"`
}

type StopInfo struct {
	State string `xml:"state,attr"`

	IdfdStop struct {
		StopName  string `xml:",chardata"`
		MatchQlty int    `xml:"matchQuality,attr"`
		StopID    int    `xml:"stopID,attr"`
	} `xml:"odvNameElem"`
}

type XmlResult struct {
	Stop       StopInfo    `xml:"itdDepartureMonitorRequest>itdOdv>itdOdvName"`
	Departures []Departure `xml:"itdDepartureMonitorRequest>itdDepartureList>itdDeparture"`
}

var (
	baseURL string
)

func main() {
	station_id := flag.String("stop", "Königsplatz", "id or (part of the) stop name")
	max_results := flag.Int("results", 5, "how many results to show")
	flag.StringVar(&baseURL, "baseurl", "http://efa.avv-augsburg.de/avv/", "base-url for EFA API")
	flag.Parse()

	endpoint := "XML_DM_REQUEST"
	params := url.Values{
		"type_dm":              {"stop"},
		"name_dm":              {*station_id},
		"useRealtime":          {"1"},
		"locationServerActive": {"1"},
		"dmLineSelection":      {"all"},
		"limit":                {strconv.Itoa(*max_results)},
		"mode":                 {"direct"},
	}

	resp, err := http.PostForm(baseURL+endpoint, params)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result XmlResult
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReader
	if err = decoder.Decode(&result); err != nil {
		panic(err)
	}
	//fmt.Printf("%+v", result)

	if result.Stop.State != "identified" {
		fmt.Println("Stop does not exist or name is not unique!")
		return
	}
	fmt.Printf("Selected stop: %s (%d)\n\n",
		result.Stop.IdfdStop.StopName,
		result.Stop.IdfdStop.StopID)

	for _, departure := range result.Departures {
		plu := ""
		if departure.Countdown != 1 {
			plu = "s"
		}

		fmt.Printf("Route %-5s due in %-2d minute%s --> %s\n",
			departure.ServingLine.Number,
			departure.Countdown,
			plu,
			departure.ServingLine.Direction)
	}

}
