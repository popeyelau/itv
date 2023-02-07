package main

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var trackUrls []string
var confs []*Group

type Group struct {
	Group    string   `yaml:"group"`
	Urls     []string `yaml:"urls"`
	Keywords string   `yaml:"keywords"`
	Track    []Track
}

func main() {

	data, err := ioutil.ReadFile("sub.yaml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	err = yaml.Unmarshal(data, &confs)
	if err != nil {
		return
	}

	trackUrls = []string{}

	wg := sync.WaitGroup{}
	wg.Add(len(confs))
	for _, conf := range confs {
		go func(conf *Group) {
			defer wg.Done()
			parseConf(conf)
		}(conf)
	}
	wg.Wait()
	merge()
}

func parseConf(conf *Group) {
	wg := sync.WaitGroup{}
	wg.Add(len(conf.Urls))
	for _, url := range conf.Urls {
		go func(url string) {
			defer wg.Done()
			parseM3u(url, conf)
		}(url)
	}
	wg.Wait()
}

func parseM3u(url string, group *Group) {
	playlist, err := Parse(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	var tracks []Track
	for _, track := range playlist.Tracks {
		if filter(track, group) && !isRequested(track.URI) {
			tracks = append(tracks, track)
		}
	}

	if len(tracks) == 0 {
		fmt.Println("no channels")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(tracks))
	for _, track := range tracks {
		go func(track Track) {
			defer wg.Done()
			trackUrls = append(trackUrls, track.URI)
			if ping(track.URI) {
				track.buildTags(group)
				group.Track = append(group.Track, track)
			}
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
	isValid := isValidRespCode(code)

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

	for _, v := range confs {
		tracks = append(tracks, v.Track...)
	}

	sort.Slice(tracks, func(i, j int) bool {
		return strings.TrimSpace(tracks[i].Name) < strings.TrimSpace(tracks[j].Name)
	})

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

func filter(track Track, conf *Group) bool {
	var name, group string
	for _, tag := range track.Tags {
		if tag.Name == "group-title" {
			group = strings.ToLower(tag.Value)
		}
	}
	name = strings.ToLower(track.Name)
	keywords := strings.Split(conf.Keywords, ",")

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
