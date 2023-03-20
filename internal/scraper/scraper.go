package scraper

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	rootPath       = "./data"
	elementPerPage = 200
	baseURL        = "https://hobbygames.ru/catalog-all"
	domain         = "hobbygames.ru"
	requestTimeout = time.Second * 30
)

func Start(ctx context.Context) {
	session, _ := newSession(ctx, rootPath)
	baseCollector := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.CacheDir(rootPath+"/cache"),
	)
	q, err := queue.New(3, &queue.InMemoryQueueStorage{MaxSize: 10000})
	if err != nil {
		log.Printf("Can't create task queue: %v", err)
	}

	// Filling the queue
	done := make(chan struct{})
	c1 := baseCollector.Clone()
	// looking for ref to the last page
	c1.OnHTML(".last", func(element *colly.HTMLElement) {
		defer close(done)
		href := element.Attr("href")
		pageCount, err := parseHref(href)
		if err != nil || pageCount == 0 {
			log.Panicf("can't plan task. Bad data after first query recieved: %v", err)
		}
		log.Printf("task params: resultsPerPage %d, pageCount %d", elementPerPage, pageCount)

		for i := 1; i <= pageCount; i++ {
			taskURL := fmt.Sprintf("%s?page=%d&results_per_page=%d&parameter_type=0", baseURL, i, elementPerPage)
			err = q.AddURL(taskURL)
			if err != nil {
				log.Panicf("can't add task to the queue: %v", err)
			}
		}
		sz, _ := q.Size()
		log.Printf("queue filled. current size %d", sz)
	})
	catalogUrl := fmt.Sprintf("%s?results_per_page=%d", baseURL, elementPerPage)
	err = c1.Visit(catalogUrl)
	if err != nil {
		log.Panicf("can't visit catalog page: %v", err)
	}
	<-done
	//

	// Start main collector
	mainCollector := baseCollector.Clone()
	mainCollector.SetRequestTimeout(requestTimeout)
	// callback functions
	saveResponseFn := func(r *colly.Response) {
		fn := session.fileName(r.Request.URL)
		err := r.Save(fn)
		if err != nil {
			log.Printf("Can't write file: %v", err)
		}
	}

	processProductItemFn := func(e *colly.HTMLElement) {
		var item Item
		item.ProductID = e.Attr("data-product_id")
		item.Price = e.Attr("data-price")

		b := e.DOM.Find("div.name-desc > a")
		if title, exists := b.Attr("title"); exists {
			title = strings.ReplaceAll(title, "\n", "")
			item.Title = title
		}

		if href, exists := b.Attr("href"); exists {
			item.HREF = href
		}
		item.Desc = strings.ReplaceAll(e.DOM.Find("div.name-desc > div.desc").Text(), "\n", "")

		b = e.DOM.Find("div.params").Children().Each(func(_ int, s *goquery.Selection) {
			if el, exists := s.Attr("class"); exists {
				val, exists := s.Attr("title")
				if !exists {
					return
				}
				val = strings.ReplaceAll(val, "\n", "")
				switch el {
				case "params__item players":
					item.NumberOfPlayers = val
				case "params__item time":
					item.GameTime = val
				case "params__item age":
					item.Age = val
				}
			}
		})
		item.SrcPageRef = e.Request.URL.String()
		session.appendItem(item)
	}

	mainCollector.OnResponse(saveResponseFn)
	mainCollector.AllowURLRevisit = false

	mainCollector.OnHTML("div.product-item", processProductItemFn)
	mainCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("got error", err)
	})

	err = q.Run(mainCollector)
	if err != nil {
		log.Printf("error occurred: %v", err)
	}
}

func ClearAllData(_ context.Context) {
	err := os.RemoveAll(rootPath + "/")
	if err != nil {
		log.Printf("Can't remove data folder: %v", err)
	}
	err = os.MkdirAll(rootPath, 0777)
	if err != nil {
		log.Printf("Can't create data folder: %v", err)
	}
}

func parseHref(src string) (int, error) {
	//<a href="?page=432&results_per_page=30" class="last">
	u, err := url.Parse(src)
	if err != nil {
		return 0, err
	}
	page := u.Query().Get("page")
	res, err := strconv.Atoi(page)
	if err != nil {
		return 0, err
	}
	return res, nil
}
