package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var bingDomains = map[string]string{
	"com": "",
	"uk": "&cc=GB",
	"us": "&cc=US",
	"tr": "&cc=TR",
	"tw": "&cc=TW",
	"ch": "&cc=CH",
	"se": "&cc=SE",
	"es": "&cc=ES",
	"za": "&cc=ZA",
	"sa": "&cc=SA",
	"ru": "&cc=RU",
	"ph": "&cc=PH",
	"pt": "&cc=PT",
	"pl": "&cc=PL",
	"cn": "&cc=CN",
	"no": "&cc=NO",
	"nz": "&cc=NZ",
	"nl": "&cc=NL",
	"mx": "&cc=MX",
	"my": "&cc=MY",
	"kr": "&cc=KR",
	"jp": "&cc=JP",
	"it": "&cc=IT",
	"id": "&cc=ID",
	"in": "&cc=IN",
	"hk": "&cc=HK",
	"de": "&cc=DE",
	"fr": "&cc=FR",
	"fi": "&cc=FI",
	"dk": "&cc=DK",
	"cl": "&cc=CL",
	"ca": "&cc=CA",
	"br": "&cc=BR",
	"be": "&cc=BE",
	"at": "&cc=AT",
	"au": "&cc=AU",
	"ar": "&cc=AR",
}

type SearchResults struct {
	ResultRank int
	ResultURL string
	ResultTitle string
	ResultDesc string
}

var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func randomUserAgent() string{
	rand.Seed(time.Now().Unix())
	randNum := rand.Int()%len(UserAgents)
	return UserAgents[randNum]
}

func buildBingURL(searchTerm string, country string, pages int, count int,)([]string, error){
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm," ","+", -1)
	if countryCode , found := bingDomains[country]; found{
		for i:= 0 ;i< pages ; i++{
			first := firstParameter(i,count)
			scrapeURL := fmt.Sprintf("https://bing.com/search?q=%s&first=%d&count=%d%s",searchTerm, first, count, countryCode)
			toScrape = append(toScrape, scrapeURL)
		}
	}else{
		err := fmt.Errorf("country %s currently not supported",country)
		return nil,err
	}
	return toScrape, nil
}

func firstParameter(number int,count int) int{
	if number == 0{
		return 1
	}
	return number+count+1
}

func scrapeClientRequest(searchURL string ,proxyString interface{})(*http.Response, error){
	baseClient := getScrapeClient(proxyString)
	req, _ := http.NewRequest("GET",searchURL, nil)
	req.Header.Set("User-Agent",randomUserAgent())

	res, err:= baseClient.Do(req)
	if res.StatusCode != 200{
		err = fmt.Errorf("scraper received a non 200 status code suggesting a ban")
		return nil ,err
	}

	if err != nil{
		return nil , err
	}

	return res,nil

}

func getScrapeClient(proxyString interface{}) *http.Client{
	switch V := proxyString.(type){
	case string:
		proxyUrl, _ := url.Parse(V)
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	default : 
		return &http.Client{}
	}
} 

func BingScrape(searchTerm string, country string, proxyString interface{},pages int, count int, backoff int)([]SearchResults, error){
	results := []SearchResults{}
	bingPages,err := buildBingURL(searchTerm, country, pages, count)
	if err != nil{
		return nil, err
	}

	for _,page := range bingPages {
		rank := len(results)
		res, err := scrapeClientRequest(page,proxyString)
		if err!= nil{
			return nil,err
		}

		data,err := bingResultParser(res,rank)
		if err!= nil{
			return nil,err
		}

		for _,result := range data {
			results = append(results, result)
		}
		time.Sleep(time.Duration(backoff)*time.Second)
	}
	return results,nil
}

func bingResultParser(response *http.Response, rank int)([]SearchResults, error){
	doc , err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil , err
	}
	results := []SearchResults{}

	sel := doc.Find("li.b_algo")
	rank++

	for i := range sel.Nodes{
		item := sel.Eq(i)
		linktag := item.Find("a")
		link, _ := linktag.Attr("href")
		titleTag := item.Find("h2")
		descTag := item.Find("text")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != "" && link != "#" && !strings.HasPrefix(link, "/"){
			result := SearchResults{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func main(){

	resp,err := BingScrape("rishabh pant", "com", nil , 2 , 30, 10)
	if err == nil{
		for _,res := range resp{
			fmt.Println(res)
			fmt.Println()
		}
	}else{
		fmt.Println(err)
	}
}