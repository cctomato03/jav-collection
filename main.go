package main

import (
	"encoding/json"
	"fmt"
	"github.com/imroc/req/v3"
	"io"
	"net/http"
	"os"
	"slices"
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

const token = "Bearer eyJhbGciOiJIUzUxMiJ9.eyJ3ZWJfbG9naW5fdXNlcl9rZXkiOiJjNDRjNWRhMS05NzE0LTQ1NjEtYTJlYS0yODhmZjBjOTRmMmMifQ.OmiluPuau4fC7PPzf61uReRWvawZuDwC1hSfWqrxFfoeV-R_uueOtpQdhcLelB4wqH220Q93cy-GoOeg2Y7e0Q"
const performerId = 4429331561381959
const performerName = "lala"

var allDownloadLink = ""

func getPageList(pageNum int) ([]PageInfo, error) {
	url := fmt.Sprintf("https://www.11jav.xyz/prod-api/api/censored/list?performerId=%d&pageNum=%d&pageSize=48&sort=pubTime&modName=censored", performerId, pageNum)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get请求失败：", err)
		panic(err)
	}
	//request.Header.Add("Rsa", "o/YRVRGI9HoVG+q216Z0aLp9hfFMxtv/3E0Dyn+o1WuOET/5/frs2qMZFsuuuK1+bzZmshCTrJQ6V5UZd+TJgA==")
	//request.Header.Add("Timestamp", "1723084531285")

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

func GetDownloadLink(pageInfo PageInfo) {
	url := fmt.Sprintf("https://www.11jav.xyz/prod-api/web/search/list?pageSize=48&pageNum=1&order=dht&data=0&keyword=%s&category=video", pageInfo.Number)
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
			allDownloadLink = fmt.Sprintf("%s\n%s %s magnet:?xt=urn:btih:%s", allDownloadLink, pageInfo.PubTime, pageInfo.Number, maxLinkInfo.InfoHash)
		}
	}
}

func DownloadPhoto(pageInfoList []PageInfo) []PageInfo {
	var unPageInfoList []PageInfo

	for _, pageInfo := range pageInfoList {
		time.Sleep(time.Second * 3)
		movieInfo, _ := getMovieInfo(pageInfo.CensoredId)

		if len(movieInfo.Performer) == 1 {
			GetDownloadLink(pageInfo)
			photoUrl := fmt.Sprintf("https://images.ssssjav.com%s", pageInfo.PicBig)
			if strings.HasPrefix(pageInfo.PicBig, "http") {
				photoUrl = pageInfo.PicBig
			}
			pathName := fmt.Sprintf("D:\\%s\\%s-%s.jpg", performerName, pageInfo.PubTime, pageInfo.Number)

			client := req.C()
			if _, err := client.R().SetOutputFile(pathName).Get(photoUrl); err != nil {
				fmt.Printf("\nF-%s-%d-%s-%d", pageInfo.Number, pageInfo.CensoredId, pageInfo.PicBig, len(movieInfo.Performer))
				//unPageInfoList = append(unPageInfoList, pageInfo)
			}
			//		response, err := client.R().SetFile("file", PathName).Post("https://canting.yunshanhu.showye.tech/api/upload/image")
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
	pageNumber := 1

	var pageInfoList []PageInfo

	for {
		pageList, err := getPageList(pageNumber)
		if err != nil {
			break
		}

		if len(pageList) > 0 {
			pageInfoList = slices.Concat(pageInfoList, pageList)
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

			if len(allDownloadLink) > 0 {
				pathName := fmt.Sprintf("D:\\%s\\%s.txt", performerName, performerName)

				f, _ := os.Create(pathName)
				defer f.Close()

				_, _ = f.WriteString(allDownloadLink) //写入文件(字节数组)
				_ = f.Sync()
			}

			break
		}
	}
}
