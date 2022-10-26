package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	wbSnapshotApiURL = "https://web.archive.org/cdx/search/xd?output=json&url=%s&fl=timestamp,original&collapse=digest&gzip=false&filter=statuscode:200"
	wbFileURL        = "https://web.archive.org/web/%sid_/%s"
)

func main() {
	var url string

	flag.Parse()

	if flag.NArg() > 0 {
		url = flag.Arg(0)
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			url = sc.Text()
		}

		if err := sc.Err(); err != nil {
			log.Fatalf("failed to read input: %s\n", err)
		}
	}

	client := http.Client{
		Timeout: time.Second * 5,
	}

	snapshots, err := getSnapshots(client, url)
	if err != nil {
		log.Fatalf("failed to snapshots: %s\n", err)
	}

	lastSnapshot := snapshots[len(snapshots)-1]
	snapshotContent, err := getSnapshotContent(client, lastSnapshot[0], lastSnapshot[1])
	if err != nil {
		log.Fatalf("failed to read input: %s\n", err)
	}

	io.Copy(os.Stdout, snapshotContent)
}

func getSnapshots(c http.Client, url string) ([][]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(wbSnapshotApiURL, url), nil)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: failed to generate request waybackmachine api: %w", err)
	}

	rsp, err := c.Do(req)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: failed to send request waybackmachine api: %w", err)
	}
	defer rsp.Body.Close()

	var r [][]string
	dec := json.NewDecoder(rsp.Body)

	err = dec.Decode(&r)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: error while decoding response %w", err)
	}

	if len(r) < 1 {
		return [][]string{}, errors.New("getSnapshots: no results found for this url")
	}

	return r[1:], nil
}

func getSnapshotContent(c http.Client, ts, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(wbFileURL, ts, url), nil)
	if err != nil {
		return nil, fmt.Errorf("getSnapshotContent: failed to generate request waybackmachine api: %w", err)
	}
	req.Header.Add("Accept-Encoding", "plain")

	rsp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getSnapshotContent: failed to send request waybackmachine api: %w", err)
	}
	defer rsp.Body.Close()

	return rsp.Body, nil
}
