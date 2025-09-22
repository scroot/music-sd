package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/scroot/music-sd/models"
)

type Music struct {
	Album    string `json:"album"`
	Singer   string `json:"singer"`
	Source   string `json:"source"`
	Duration string `json:"duration"`
	Title    string `json:"title"`
	ID       int    `json:"id"`
	Size     string `json:"size"`
	MID      string `json:"mid"`
	Url      string `json:"url"`
	Rate     string `json:"rate"`
	Name     string `json:"name"`
}

func (m Music) ReadCloser() (io.ReadCloser, error) {
	response, err := http.Get(m.Url)
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}

func (m Music) Get(filename string) (err error) {
	if filename == "" {
		//fmt.Println(m.Name)
		filename = m.Name
	}
	if runtime.GOOS == "windows" {
		compile, err := regexp.Compile("[\\/:*?\"<>|]")
		if err != nil {
			log.Panic(err)
		}
		filename = compile.ReplaceAllString(filename, ",")
	}

	//fmt.Println("开始下载", filename)
	t1 := time.Now()

	bodyReadCloser, err := m.ReadCloser()
	if err != nil {
		return err
	}

	defer func() {
		err = bodyReadCloser.Close()
		return
	}()

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(file, bodyReadCloser)
	if err != nil {
		return err
	}

	elapsed := time.Since(t1)
	fmt.Printf("%s 下载完成, 音质%skbps, 耗时: %v\n", filename, m.Rate, elapsed)
	return nil
}

func (m *Music) ParseMusic() {
	switch m.Source {
	case "NETEASE":
		m.parseNeteaseSource()
	case "QQ":
		m.parseQQSource()
	}
}

func (m *Music) parseQQSource() {
	// 根据songmid等信息获得下载链接
	guid := Random(100000000, 10000000000)
	req, err := http.NewRequest("GET", "http://base.music.qq.com/fcgi-bin/fcg_musicexpress.fcg", nil)
	AddHeader(req)
	req.Header.Set("referer", "http://m.y.qq.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1")

	q := req.URL.Query()
	q.Add("guid", strconv.Itoa(guid))
	q.Add("format", "json")
	q.Add("json", "3")

	req.URL.RawQuery = q.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}
	//fmt.Printf("%s\n", content)

	respJson := make(map[string]interface{})
	err = json.Unmarshal(content, &respJson)
	if err != nil {
		log.Panic(err)
	}

	vkey := respJson["key"]
	prefixs := []string{"M800", "M500", "C400"}
	for _, prefix := range prefixs {
		url := fmt.Sprintf("http://dl.stream.qqmusic.qq.com/%v%v.mp3?vkey=%v&guid=%v&fromtag=1", prefix, m.MID, vkey, guid)
		//fmt.Println(url)
		size := GetContentLen(url)
		//	mSize, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(size)/1048576), 64)
		mSize := fmt.Sprintf("%.2f", float64(size)/1048576)
		if size > 0 {
			m.Url = url
			if prefix == "M800" {
				m.Rate = "320"
			} else {
				m.Rate = "128"
			}
			m.Size = mSize
			break
		}
	}
	m.Name = fmt.Sprintf("%v - %v.mp3", m.Singer, m.Title)
}

func (m *Music) parseNeteaseSource() {
	musicId := "[" + strconv.Itoa(m.ID) + "]"
	// 初始化requestJson
	requestJSON := map[string]interface{}{
		"method": "POST",
		"url":    "http://music.163.com/api/song/enhance/player/url",
		"params": map[string]interface{}{
			"ids": musicId,
			"br":  320000,
		},
	}

	// json化数据
	requestBytes, err := json.Marshal(requestJSON)
	if err != nil {
		log.Panic(err)
	}
	encryptedString := EncryptForm(requestBytes)

	// post form
	form := url.Values{}
	form.Add("eparams", encryptedString)

	req, err := http.NewRequest("POST", "http://music.163.com/api/linux/forward", strings.NewReader(form.Encode()))

	// FAKE_HEADERS
	AddHeader(req)
	req.Header.Set("referer", "http://music.163.com/")

	client := http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Panic(err)
	}

	//fmt.Println(resp.StatusCode)
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err.Error())
	}
	//fmt.Printf("content: %s\n", content)

	var respJson models.MusicDownloadNetease

	err = json.Unmarshal(content, &respJson)
	if err != nil {
		log.Panic(err)
	}

	//fmt.Println(respJson.Code)
	if respJson.Code != 200 {
		log.Panic("code not 200 ", respJson.Code)
	}

	m.Url = respJson.Data[0].URL
	m.Name = fmt.Sprintf("%v - %v.%v", m.Singer, m.Title, respJson.Data[0].Type)
	m.Rate = strconv.Itoa(respJson.Data[0].Br / 1000)
}
