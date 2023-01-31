package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var groups map[string][]Track
var trackUrls []string
var keywords []string

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
				groups[url] = append(groups[url], track)
			}
			wg.Done()
		}(track)
	}
	wg.Wait()
}

func ping(url string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	isValid := resp.StatusCode >= 200 && resp.StatusCode <= 299 || resp.StatusCode == 302
	return isValid
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
