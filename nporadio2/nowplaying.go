package nporadio2

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
	"os"
)

type NowPlayingSongRoot struct {
	Results []NowPlayingSong `json:"results"`
}

type NowPlayingSong struct {
	Id            uint64                 `json:"id"`
	Startdatetime string                 `json:"startdatetime"`
	Stopdatetime  string                 `json:"stopdatetime"`
	Date          string                 `json:"date"`
	Songfile      NowPlayingSongSongFile `json:"songfile"`
	Channel       int8                   `json:"channel"`
}

type NowPlayingSongSongFile struct {
	Id            uint64                    `json:"id"`
	Artist        string                    `json:"artist"`
	Title         string                    `json:"title"`
	DaletId       string                    `json:"dalet_id"`
	SongId        uint64                    `json:"song_id"`
	Hidden        int8                      `json:"hidden"`
	LastUpdated   string                    `json:"last_updated"`
	BumaId        string                    `json:"buma_id"`
	Rb1Id         uint64                    `json:"rb1id"`
	SongVersion   NowPlayingSongFileVersion `json:"songversion"`
	References    NowPlayingSongReference   `json:"_references"`
	ReferencesSsl NowPlayingSongReference   `json:"_references_ssl"`
}

type NowPlayingSongFileVersion struct {
	Id    uint64                `json:"id"`
	Image []NowPlayingSongImage `json:"image"`
}

type NowPlayingSongImage struct {
	Id             uint64 `json:"id"`
	Name           string `json:"name"`
	Filename       string `json:"filename"`
	Hash           string `json:"hash"`
	Source         string `json:"source"`
	OriginalWidth  uint32 `json:"original_width"`
	OriginalHeight uint32 `json:"original_height"`
	Deleted        uint8  `json:"deleted"`
	Created        string `json:"created"`
	AllowedToUse   uint8  `json:"allowed_to_use"`
	Replaced       uint8  `json:"replaced"`
	Updated        string `json:"updated"`
	Url            string `json:"url"`
	UrlSsl         string `json:"url_ssl"`
}

type NowPlayingSongReference struct {
	Channel string `json:"channel"`
}

type QueryStruct struct {
	NpoCcSkipWall int `url:npo_cc_skip_wall`
}

type NpoError struct {
}

func newNowPlayingService() *NowPlayingService {
	return &NowPlayingService{
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
		sling:     sling.New().Base("http://radiobox2.omroep.nl/data/radiobox2/nowonair/"),
		//?npo_cc_skip_wall=1
	}
}

type NowPlayingService struct {
	UserAgent string
	sling     *sling.Sling
	lastSongs []*NowPlayingSong
	log       *logrus.Logger
}

func (n *NowPlayingService) addLogger(log *logrus.Logger) {
	n.log = log
}

func (n *NowPlayingService) WarmupCache() []*NowPlayingSong {

	if(len(n.lastSongs) > 0){
		n.lastSongs = make([]*NowPlayingSong, 0)
	}

	files, err := ioutil.ReadDir("cache/songs")
	if err != nil {
		n.log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			content, _ := ioutil.ReadFile("cache/songs/" + file.Name())
			cachedSong := new(NowPlayingSong)
			json.Unmarshal(content, &cachedSong)
			n.lastSongs = append(n.lastSongs, cachedSong)
		}
	}
	return n.lastSongs
}

func (n *NowPlayingService) GetPlayedSongs() []*NowPlayingSong {
	return n.lastSongs
}

func (n *NowPlayingService) isNewSong(song *NowPlayingSong) bool {

	n.log.Debugf("Existing songs count is: %d", len(n.lastSongs))
	if len(n.lastSongs) == 0 {
		n.WarmupCache()
		n.log.Debugf("Cache warm up, Existing songs count is updated to: %d", len(n.lastSongs))
	}

	for _, currentSong := range n.lastSongs{
			if song.Songfile.Title == currentSong.Songfile.Title && song.Songfile.Artist == currentSong.Songfile.Artist {
				n.log.Debugf("Found current song based on Title %s and Artist %s.", song.Songfile.Title, song.Songfile.Artist)
				return false
			}
			//if song.Id == currentSong.Id {
			//	n.log.Debugf("Found current song based on Id: %i.", song.Id)
			//	return false
			//}
	}

	jsonStructureBytes, _ := json.MarshalIndent(song, "", "\t")
	ioutil.WriteFile(fmt.Sprintf("cache/songs/%d.json", time.Now().Unix()), jsonStructureBytes, 0644)
	n.log.Debug("Found a new song.")
	n.lastSongs = append(n.lastSongs, song)
	return true
}

func (n *NowPlayingService) NextSongInSeconds(song *NowPlayingSong) int64 {
	layout := "2006-01-02T15:04:05-07:00"
	endTime, err := time.Parse(layout, song.Stopdatetime)
	location, _ := time.LoadLocation("Europe/Amsterdam")
	curTime := time.Now().In(location)

	if err != nil {
		fmt.Printf("String: %s", err.Error())
	}

	endTime = time.Unix(endTime.Unix()+7200, 0)

	nextIn := endTime.Unix() - curTime.Unix()

	fmt.Printf("STR Time: %s\n", song.Stopdatetime)

	t := endTime
	fmt.Printf("END Time: %d-%02d-%02dT%02d:%02d:%02d\n",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	t = curTime
	t = curTime
	fmt.Printf("NOW Time: %d-%02d-%02dT%02d:%02d:%02d\n",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	//return diff.Seconds()
	return nextIn
}

func (n *NowPlayingService) CurrentSong() (*NowPlayingSong, bool) {

	songRoot := new(NowPlayingSongRoot)
	error := new(NpoError)

	n.sling.Get("2.json").QueryStruct(QueryStruct{NpoCcSkipWall: 1}).Receive(songRoot, error)

	if len(songRoot.Results) > 0 {
		song := &songRoot.Results[0]
		return song, n.isNewSong(song)
	} else {
		n.log.Error("Empty results set received.")
		jsonStructureBytes, _ := json.MarshalIndent(songRoot, "", "\t")
		ioutil.WriteFile(fmt.Sprintf("cache/errors/%d.json", time.Now().Unix()), jsonStructureBytes, 0644)
		os.Exit(3)
	}
	return nil, false
}
