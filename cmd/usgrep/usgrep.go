/*
usgrep is the Sonar UDP OpenData grep tool
*/
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type record struct {
	Timestamp       time.Time
	SourceAddr      net.IP
	SourcePort      int
	DestinationAddr net.IP
	DestinationPort int
	IPID            int
	TTL             int
	Data            string
}

func unmarshal(s string) (*record, error) {
	tokens := strings.Split(s, ",")
	if len(tokens) != 8 {
		return nil, fmt.Errorf("invalid line: %s", s)
	}

	ts, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %s", tokens[0])
	}

	sp, err := strconv.Atoi(tokens[2])
	if err != nil {
		return nil, fmt.Errorf("invalid source port fromat: %s", tokens[2])
	}

	dp, err := strconv.Atoi(tokens[4])
	if err != nil {
		return nil, fmt.Errorf("invalid destination port fromat: %s", tokens[4])
	}

	ipid, err := strconv.Atoi(tokens[5])
	if err != nil {
		return nil, fmt.Errorf("invalid ipid fromat: %s", tokens[5])
	}

	ttl, err := strconv.Atoi(tokens[6])
	if err != nil {
		return nil, fmt.Errorf("invalid ttl fromat: %s", tokens[6])
	}

	return &record{
		Timestamp:       time.Unix(ts, 0),
		SourceAddr:      net.ParseIP(tokens[1]),
		SourcePort:      sp,
		DestinationAddr: net.ParseIP(tokens[3]),
		DestinationPort: dp,
		IPID:            ipid,
		TTL:             ttl,
		Data:            tokens[7],
	}, nil
}

func getSubnets(f string) ([]*net.IPNet, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	var nets []*net.IPNet
	for sc.Scan() {
		_, n, err := net.ParseCIDR(sc.Text())
		if err != nil {
			return nets, err
		}

		nets = append(nets, n)
	}

	return nets, nil
}

func main() {
	subnetsFile := flag.String("i", "", "Filename with subnets of interest inside, new line separated")
	flag.Parse()

	var subnets []*net.IPNet
	var err error
	if *subnetsFile != "" {
		subnets, err = getSubnets(*subnetsFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	g, err := gzip.NewReader(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	ch := make(chan string, 1000)

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		proc(ch, subnets, &wg)
	}()

	breader := bufio.NewReader(g)
	for {
		line, err := breader.ReadString('\n')
		if err == nil {
			ch <- line
			continue
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			close(ch)
			break
		}

		log.Println(err)
	}
	wg.Wait()
}

func dump(r *record) {
	log.Printf("%v\n", r)
}

func proc(in chan string, ss []*net.IPNet, wg *sync.WaitGroup) {
	for line := range in {
		r, err := unmarshal(line)
		if err != nil {
			log.Printf("wrong entry: %s", line)
			continue
		}

		for i := range ss {
			if ss[i].Contains(r.SourceAddr) {
				dump(r)
				break
			}
		}
	}
}
