package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var s sync.RWMutex
var groups map[string][]Track
var trackUrls []string
var keywords []string

func main() {

	files := loadSource("source.txt")
	keywords = loadSource("keywords.txt")

	groups = map[string][]Track{}
	trackUrls = []string{}

	wg := sync.WaitGroup{}
	wg.Add(len(files))
	for _, url := range files {
		groups[url] = []Track{}
		go func(url string) {
			parse(url)
			wg.Done()
		}(url)
	}
	wg.Wait()
	merge()
}

func parse(url string) {
	playlist, err := Parse(url)
	if err != nil {
		return
	}

	var tracks []Track
	for _, track := range playlist.Tracks {
		if filter(track) && !isRequested(track.URI) {
			tracks = append(tracks, track)
		}
	}

	if len(tracks) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(tracks))
	for _, track := range tracks {
		go func(track Track) {
			trackUrls = append(trackUrls, track.URI)
			if ping(track.URI) {
				s.Lock()
				groups[url] = append(groups[url], track)
				s.Unlock()
			}
			wg.Done()
		}(track)
	}
	wg.Wait()
}

func ping(url string) bool {
	resp, err := request(url, "HEAD", 3*time.Second, true)
	if err != nil {
		return false
	}

	code := resp.StatusCode
	isRedirect := code == 302
	isValid := isValidRespCode(code)

	if isRedirect {
		location := resp.Header.Get("Location")
		if len(location) == 0 {
			return false
		}

		resp, err = request(location, "HEAD", 3*time.Second, false)
		if err != nil {
			return false
		}
		isValid = isValidRespCode(resp.StatusCode)
	}

	return isValid
}

func request(url string, method string, timeout time.Duration, checkRedirect bool) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if checkRedirect {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func merge() {
	var tracks []Track

	for _, v := range groups {
		if len(v) == 0 {
			continue
		}
		tracks = append(tracks, v...)
	}

	playlist := Playlist{Tracks: tracks}
	reader, err := Marshall(playlist)
	if err != nil {
		fmt.Println(err)
		return
	}
	b := reader.(*bytes.Buffer)
	_ = ioutil.WriteFile("./merged.m3u", b.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func filter(track Track) bool {
	var name, group string
	for _, tag := range track.Tags {
		if tag.Name == "group-title" {
			group = strings.ToLower(tag.Value)
		}
	}
	name = strings.ToLower(track.Name)

	for _, keyword := range keywords {
		if strings.Contains(name, keyword) || strings.Contains(group, keyword) {
			return true
		}
	}
	return false
}

func isRequested(url string) bool {
	for _, v := range trackUrls {
		if v == url {
			return true
		}
	}
	return false
}

func isValidRespCode(statusCode int) bool {
	return (statusCode >= 200 && statusCode < 300) || statusCode == 302
}

func loadSource(path string) []string {
	readFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var files []string

	for fileScanner.Scan() {
		files = append(files, fileScanner.Text())
	}

	_ = readFile.Close()
	return files
}
