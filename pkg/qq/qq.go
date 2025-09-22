package qq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/scroot/music-sd/models"
	"github.com/scroot/music-sd/pkg/common"
)

func Search(keyword string) (musicList []common.Music) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "http://c.y.qq.com/soso/fcgi-bin/search_for_qq_cp", nil)
	common.AddHeader(req)
	req.Header.Set("referer", "http://m.y.qq.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1")

	q := req.URL.Query()
	q.Add("w", keyword)
	q.Add("format", "json")
	q.Add("p", "1")
	// count
	q.Add("n", "8")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}
	var respJson models.RespQQ
	err = json.Unmarshal(content, &respJson)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("%s\n", content)
	if respJson.Code != 0 {
		log.Panic("code isn't 0", respJson.Code)
	}

	var music common.Music
	for _, song := range respJson.Data.Song.List {
		// singer
		var singers []string
		for _, singer := range song.Singer {
			singers = append(singers, singer.Name)
		}

		size := song.Size128
		if song.Size320 != 0 {
			size = song.Size320
		}
		mSize := fmt.Sprintf("%.2f", float64(size)/1048576)

		music.Title = song.Songname
		music.ID = song.Songid
		music.MID = song.Songmid
		music.Duration = time.Unix(int64(song.Interval), 0).Format("04:05")
		music.Singer = strings.Join(singers, ",")
		music.Album = song.Albumname
		music.Size = mSize
		music.Source = "QQ"
		musicList = append(musicList, music)
	}
	return musicList
}
