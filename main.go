package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/sfomuseum/go-exif-update"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type LinkInfo struct {
	InfoHash string `json:"infoHash"`
	Length   int64  `json:"length"`
}

type DownloadLinkInfo struct {
	Rows []LinkInfo `json:"rows"`
}

type PageListResult struct {
	PageInfo []PageInfo `json:"rows"`
	Total    int        `json:"total"`
}

type PageInfo struct {
	CensoredId int64  `json:"censoredId"`
	Number     string `json:"number"`
	PicBig     string `json:"picBig"`
	PubTime    string `json:"pubTime"`
	Title      string `json:"title"`
	Count      int    `json:"count"`
}

type MovieActress struct {
	PerformerId     int64  `json:"performerId"`
	PerformerCnName string `json:"performerCnName"`
	PerformerAvatar string `json:"performerAvatar"`
}

type MovieData struct {
	CensoredId int64  `json:"censoredId"`
	PicBig     string `json:"picBig"`
	Number     string `json:"number"`
}

type MovieInfo struct {
	Performer []MovieActress `json:"performer"`
	Data      MovieData      `json:"data"`
}

// token认证
var token = "Bearer eyJhbGciOiJIUzUxMiJ9.eyJ3ZWJfbG9naW5fdXNlcl9rZXkiOiIzMmMzOTIwZS1hZTQ0LTQ1ZDUtODU5ZC0zZDk4NDczNjUwZDQifQ.qA_KQ6IYcaPJ8QkIJNb0wmNsVPtO9PpOy72fXQnZK-W1xVmebLhRTVG6QBAL3V76KvIO-_O-Wy4cl6Z5W9IggA"

// 演员ID
var performerId int64 = 4421250301886541

// 截至日期
var dealLine = "2024-07-21"

// 下载目录
var photoDir = "D:\\jav"

// 获取列表
func getPageList(pageNum int) ([]PageInfo, error) {
	url := fmt.Sprintf("https://www.11jav.xyz/prod-api/api/censored/list?performerId=%d&pageNum=%d&pageSize=48&sort=pubTime&modName=censored", performerId, pageNum)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get请求失败：", err)
		panic(err)
	}

	request.Header.Add("authorization", token)

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	result := string(b)

	var pageListResult PageListResult
	errs := json.Unmarshal([]byte(result), &pageListResult)
	if errs != nil {
		return nil, errs
	}

	return pageListResult.PageInfo, nil
}

// 获取影片信息
func getMovieInfo(censoredId int64) (*MovieInfo, error) {
	url := fmt.Sprintf("https://www.11jav.xyz/prod-api/api/censored/%d", censoredId)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get请求失败：", err)
		panic(err)
	}
	request.Header.Add("authorization", token)

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	result := string(b)

	var movieInfo MovieInfo
	errs := json.Unmarshal([]byte(result), &movieInfo)
	if errs != nil {
		return nil, errs
	}

	return &movieInfo, nil
}

// GetDownloadLink 获取下载链接
func GetDownloadLink(number string, filePath string) {
	url := fmt.Sprintf("https://www.11jav.xyz/prod-api/web/search/list?pageSize=48&pageNum=1&order=dht&data=0&keyword=%s&category=video", number)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get请求失败：", err)
		panic(err)
	}
	request.Header.Add("authorization", token)

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	result := string(b)

	var downloadLinkInfo DownloadLinkInfo
	errs := json.Unmarshal([]byte(result), &downloadLinkInfo)
	if errs == nil {
		if len(downloadLinkInfo.Rows) > 0 {
			maxLinkInfo := downloadLinkInfo.Rows[0]
			for _, row := range downloadLinkInfo.Rows {
				if maxLinkInfo.Length < row.Length {
					maxLinkInfo = row
				}
			}
			downloadUrl := fmt.Sprintf("magnet:?xt=urn:btih:%s", maxLinkInfo.InfoHash)

			writePhotoDownloadLink(downloadUrl, filePath)
		}
	}
}

func writePhotoDownloadLink(downloadLink string, filePath string) {
	exifProps := map[string]interface{}{
		"Artist": downloadLink,
	}

	source, _ := os.Open(filePath)
	defer source.Close()
	bakFilePath := fmt.Sprintf("%s.bak", filePath)
	out, _ := os.Create(bakFilePath)
	defer out.Close()

	err := update.PrepareAndUpdateExif(source, out, exifProps)
	if err != nil {
		return
	} else {
		go func() {
			time.Sleep(3 * time.Second)
			err = os.Remove(filePath)
			if err != nil {
				return
			}

			err = os.Rename(bakFilePath, filePath)
			if err != nil {
				return
			}
		}()
	}
}

// DownloadPhoto 下载图片
func DownloadPhoto(pageInfoList []PageInfo) []PageInfo {
	var unPageInfoList []PageInfo

	for _, pageInfo := range pageInfoList {
		time.Sleep(time.Second * 3)
		movieInfo, _ := getMovieInfo(pageInfo.CensoredId)

		if len(movieInfo.Performer) == 1 {
			photoUrl := fmt.Sprintf("https://images.ssssjav.com%s", pageInfo.PicBig)
			if strings.HasPrefix(pageInfo.PicBig, "http") {
				photoUrl = pageInfo.PicBig
			}
			pathName := fmt.Sprintf("%s\\%s-%s.jpg", photoDir, pageInfo.PubTime, pageInfo.Number)

			client := req.C()
			if _, err := client.R().SetOutputFile(pathName).Get(photoUrl); err != nil {
				fmt.Printf("\nF-%s-%d-%s-%d", pageInfo.Number, pageInfo.CensoredId, pageInfo.PicBig, len(movieInfo.Performer))
			} else {
				fmt.Printf("\n%s-%s", pageInfo.PubTime, pageInfo.Number)
				GetDownloadLink(pageInfo.Number, pathName)
			}
		} else if len(movieInfo.Performer) == 0 {
			unPageInfoList = append(unPageInfoList, pageInfo)
		}
		//else {
		//	fmt.Printf("\nM-%s-%d-%s-%d", pageInfo.Number, pageInfo.CensoredId, pageInfo.PicBig, len(movieInfo.Performer))
		//}
	}

	return unPageInfoList
}

func main() {
	flag.StringVar(&token, "token", "Bearer eyJhbGciOiJIUzUxMiJ9.eyJ3ZWJfbG9naW5fdXNlcl9rZXkiOiIzMmMzOTIwZS1hZTQ0LTQ1ZDUtODU5ZC0zZDk4NDczNjUwZDQifQ.qA_KQ6IYcaPJ8QkIJNb0wmNsVPtO9PpOy72fXQnZK-W1xVmebLhRTVG6QBAL3V76KvIO-_O-Wy4cl6Z5W9IggA", "login token")
	// 演员ID
	flag.Int64Var(&performerId, "performerId", 4421250301886541, "actress id")
	// 截至日期
	flag.StringVar(&dealLine, "date", "2024-07-21", "dead-line")
	// 下载目录
	flag.StringVar(&photoDir, "photoDir", "D:\\jav", "photo dir")
	flag.Parse()

	pageNumber := 1

	var pageInfoList []PageInfo

	isDeadLine := false

	for {
		pageList, err := getPageList(pageNumber)
		if err != nil {
			break
		}

		if len(pageList) > 0 {
			for _, pageInfo := range pageList {
				pubDate, pubErr := time.Parse("2006-01-02", pageInfo.PubTime)
				dealLineDate, deadLineErr := time.Parse("2006-01-02", dealLine)
				if pubErr != nil || deadLineErr != nil || pubDate.Before(dealLineDate) {
					isDeadLine = true
					break
				}
				pageInfoList = append(pageInfoList, pageInfo)
			}
		}

		if isDeadLine {
			break
		}

		if len(pageList) < 48 {
			break
		}
		//break
		pageNumber++
		time.Sleep(time.Second * 3)
	}

	fmt.Printf("\ntotal movie is %d", len(pageInfoList))

	for {
		unPageInfoList := DownloadPhoto(pageInfoList)

		if len(unPageInfoList) > 0 {
			pageInfoList = unPageInfoList
		} else {
			fmt.Printf("\nall download end")
			break
		}
	}
}
