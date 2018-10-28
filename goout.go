package main

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

type NewsMap struct {
	Keyword string
	Location string
}

type NewsAggPage struct {
	Title string
	News map[string]NewsMap
}

type SitemapIndex struct{
	Locations []string `xml:"sitemap>loc"`
}
type News struct {
	Titles []string `xml:"url>news>title"`
	Keywords []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type IndexPage struct {
	Title string
}

func indexHandler(w http.ResponseWriter, r *http.Request){
	page :=IndexPage{Title: "landing"}
	t, _ := template.ParseFiles("./html/index.html")
	t.Execute(w, page)
}

func newsRoutine(c chan News, Location string){
	defer wg.Done()
	var n News
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()

	c <- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request){
	var s SitemapIndex
	resp, _ := http.Get("http://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	news_map := make(map[string]NewsMap)
	resp.Body.Close()

	queue := make(chan News, 30)
	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}
	wg.Wait()
	close(queue)

	for elem := range queue{
		for idx, _ := range elem.Keywords {
		news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx],elem.Locations[idx]}
		}
	}
	nPage:= NewsAggPage{Title: "Title",News: news_map}
	t, _ := template.ParseFiles("./html/newsTemplate.html")
	t.Execute(w, nPage)
	}

func main()  {
	http.HandleFunc("/",indexHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/agg/",newsAggHandler)
	http.ListenAndServe(":8000",nil)

}