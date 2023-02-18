package youtube

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	m "itv/merge"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/etherlabsio/go-m3u8/m3u8"
	"gopkg.in/yaml.v2"
)

var conf []m.Channnel

func Run() {
	data, err := ioutil.ReadFile("youtube.yaml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	playlist := m.Playlist{
		Tracks: []m.Track{},
	}

	wg.Add(len(conf))
	for _, c := range conf {
		go func(c m.Channnel) {
			defer wg.Done()
			part, _ := url.Parse(c.Url)
			stream := getLiveUrl(c.Url, c.Resolution)
			if len(stream) < 1 {
				return
			}

			playlist.Tracks = append(playlist.Tracks, m.Track{
				Name:   c.Name,
				Length: -1,
				URI:    stream,
				Tags: []m.Tag{{
					Name:  "tvg-logo",
					Value: fmt.Sprintf("https://i.ytimg.com/vi/%s/hq720_live.jpg", part.Query().Get("v")),
				}},
			})
		}(c)
	}
	wg.Wait()

	reader, err := m.Marshall(playlist)
	if err != nil {
		fmt.Println(err)
		return
	}
	b := reader.(*bytes.Buffer)
	_ = ioutil.WriteFile("./youtube.m3u", b.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func getLiveUrl(url, resolution string) string {
	client := &http.Client{
		Timeout: time.Second * 5,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	r, _ := http.NewRequest("GET", url, nil)
	r.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	resp, err := client.Do(r)
	if err != nil {
		return ""
	}

	body, _ := io.ReadAll(resp.Body)
	str := string(body)
	defer resp.Body.Close()

	reg := regexp2.MustCompile(`(?<=hlsManifestUrl":").*\.m3u8`, regexp2.RE2)
	res, _ := reg.FindStringMatch(str)
	if res == nil {
		return ""
	}
	stream := res.Captures[0].String()
	quality := getResolution(stream, resolution)
	if quality != nil {
		return *quality
	}
	return stream
}

func getResolution(liveurl string, quality string) *string {
	client := &http.Client{Timeout: time.Second * 5}
	r, _ := http.NewRequest("GET", liveurl, nil)
	r.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	resp, err := client.Do(r)
	if err != nil {
		return nil
	}
	playlist, err := m3u8.Read(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil
	}

	size := playlist.ItemSize()

	if size < 1 {
		return nil
	}

	mapping := map[string]string{}
	for _, item := range playlist.Playlists() {
		mapping[strconv.Itoa(item.Resolution.Height)] = item.URI
	}

	if stream, ok := mapping[quality]; ok {
		return &stream
	}

	stream := &playlist.Playlists()[size-1].URI
	return stream
}
