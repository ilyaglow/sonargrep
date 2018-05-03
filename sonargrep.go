package main

import (
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

// Response represents a single json record from Sonar http/https dataset
type Response struct {
	Data    string  `json:"data"`
	Host    string  `json:"host"`
	IP      string  `json:"ip"`
	Path    string  `json:"path"`
	Port    int     `json:"port"`
	Subject Subject `json:"subject,omitempty"` // https only
	VHost   string  `json:"vhost"`
}

// Subject represents TLS headers
type Subject struct {
	C                string `json:"C,omitempty"`
	CN               string `json:"CN,omitempty"`
	BusinessCategory string `json:"businessCategory,omitempty"`
	JurisdictionST   string `json:"jurisdictionST,omitempty"`
	SerialNumber     string `json:"serialNumber,omitempty"`
	L                string `json:"L,omitempty"`
	O                string `json:"O,omitempty"`
	ST               string `json:"ST,omitempty"`
	Street           string `json:"street,omitempty"`
	JurisdictionL    string `json:"jurisdictionL,omitempty"`
	PostalCode       string `json:"postalCode,omitempty"`
	OU               string `json:"OU,omitempty"`
	JurisdictionC    string `json:"jurisdictionC,omitempty"`
}

func main() {
	gword := flag.String("w", "", "word to grep")
	casesens := flag.Bool("i", true, "ignore case")
	flag.Parse()

	g, err := gzip.NewReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	chl := make(chan []byte, 1000)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		proc(chl, *gword, *casesens, &wg)
	}()

	breader := bufio.NewReader(g)
	for {
		line, err := breader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				close(chl)
				break
			}
			if err == io.ErrUnexpectedEOF {
				close(chl)
				break
			}
			log.Println(err)
			continue
		}
		chl <- line
	}
	wg.Wait()
}

func proc(in chan []byte, w string, c bool, wg *sync.WaitGroup) {
	for line := range in {
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			log.Println(err)
			continue
		}

		data, err := base64.StdEncoding.DecodeString(resp.Data)
		if err != nil {
			log.Println(err)
			continue
		}

		if !contains(string(data), w, c) {
			continue
		}

		resp.Data = string(data)

		output, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(string(output))

	}
	wg.Done()
}

func contains(s string, w string, c bool) bool {
	switch c {
	case false:
		m := search.New(language.English, search.IgnoreCase)
		start, _ := m.IndexString(s, w)
		if start != -1 {
			return true
		}
		return false

	default:
		return strings.Contains(s, w)
	}
}
