package nporadio2

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Top2000List struct {
	Data []Top2000ListSongEntry `json:"data"`
}

type Top2000ListSongEntry struct {
	Title              string `json:"s"`
	Artist             string `json:"a"`
	ArtistID           string `json:"aid"`
	Year               int    `json:"yr"`
	Position           int    `json:"pos"`
	PreviousPosition   int    `json:"prv"`
	Image              string `json:"img"`
	Url                string `json:"url"`
	Spotify            string `json:"spotify"`
	PositionChange     int
	PositionChangeText string
}

func newTop2000Service() *Top2000Service {
	return &Top2000Service{
		CurrentPosition: 2000,
	}
}

type Top2000Service struct {
	CurrentPosition int16
	log             *logrus.Logger
	chart           *Top2000List
	client          *Client
}

func (t *Top2000Service) addLogger(log *logrus.Logger) {
	t.log = log
}

func (t *Top2000Service) addClient(client *Client) {
	t.client = client
}

func (t *Top2000Service) ReadList(location string) {
	charStay := "●"
	Chardown := "▼"
	CharUp := "▲"
	CharNew := "■"

	data, _ := ioutil.ReadFile(location)
	chart := new(Top2000List)
	json.Unmarshal(data, chart)
	t.chart = chart
	for index, chartSong := range t.chart.Data {
		chartSong.PositionChange = chartSong.Position - chartSong.PreviousPosition
		if chartSong.PreviousPosition == 0 {
			t.chart.Data[index].PositionChangeText = fmt.Sprintf("%s", CharNew)
		} else if chartSong.PositionChange == 0 {
			t.chart.Data[index].PositionChangeText = fmt.Sprintf("%s", charStay)
		} else if chartSong.PositionChange > 0 {
			t.chart.Data[index].PositionChangeText = fmt.Sprintf("%s%d", Chardown, chartSong.PositionChange)
		} else {
			t.chart.Data[index].PositionChangeText = fmt.Sprintf("%s%d", CharUp, (chartSong.PositionChange * -1))
		}
	}
}

func (t *Top2000Service) FindPosition(song *NowPlayingSong) *Top2000ListSongEntry {
	for _, chartSong := range t.chart.Data {
		if song.Songfile.Artist == chartSong.Artist && song.Songfile.Title == chartSong.Title {
			t.log.Debugf("Found based on title and artist")
			return &chartSong
		}


		if len(song.Songfile.SongVersion.Image) > 0 && fmt.Sprintf("%d", song.Songfile.SongVersion.Image[0].Id) == chartSong.Image {
			t.log.Debugf("Found based on image id: %d", song.Songfile.SongVersion.Image[0].Id)
			return &chartSong
		}
	}

	if len(t.client.NowPlaying.lastSongs) == 0 {
		t.client.NowPlaying.WarmupCache()
	}

	return &t.chart.Data[(2000 - len(t.client.NowPlaying.lastSongs))]
}


