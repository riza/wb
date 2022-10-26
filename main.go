package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
			fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
		}
	}

	client := http.Client{
		Timeout: time.Second * 5,
	}

	snapshots, err := getSnapshots(client, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to getting snapshots  %s\n", err)
	}

	lastSnapshot := snapshots[len(snapshots)-1]
	lastSnapshotTs := lastSnapshot[0]
	lastSnapshotURL := lastSnapshot[1]

	request, err := http.NewRequest("GET", fmt.Sprintf(wbFileURL, lastSnapshotTs, lastSnapshotURL), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate request waybackmachine api: %s\n", err)
	}

	request.Header.Add("Accept-Encoding", "plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request waybackmachine api: %s\n", err)
	}

	defer response.Body.Close()

	io.Copy(os.Stdout, response.Body)
}

func getSnapshots(c http.Client, url string) ([][]string, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf(wbSnapshotApiURL, url), nil)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: failed to generate request waybackmachine api: %w", err)
	}

	response, err := c.Do(request)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: failed to send request waybackmachine api: %w", err)
	}
	defer response.Body.Close()

	var r [][]string
	dec := json.NewDecoder(response.Body)

	err = dec.Decode(&r)
	if err != nil {
		return [][]string{}, fmt.Errorf("getSnapshots: error while decoding response %w", err)
	}

	if len(r) < 1 {
		return [][]string{}, errors.New("getSnapshots: no results found for this url")
	}

	return r[1:], nil
}
