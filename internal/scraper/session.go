package scraper

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	DDMMYYYYhhmmss = "2006-01-02_150405"
	flushInterval  = time.Second * 5
	firstPage      = 1
)

type session struct {
	sessionMu sync.RWMutex
	fileNum   uint
	dataPath  string
	items     []Item
	page      int
}

func newSession(ctx context.Context, root string) (*session, error) {
	var target session

	_, err := os.Stat(root)
	if err != nil {
		log.Printf("Root is not a folder: %v", err)
		return nil, err
	}

	dataPath := root + "/" + time.Now().Round(time.Second).Format(DDMMYYYYhhmmss)
	err = os.Mkdir(dataPath, 0777)
	if err != nil {
		log.Printf("Can't create data dir: %v", err)
		return nil, err
	}

	target.dataPath = dataPath
	log.Printf("Data folder is: %s", target.dataPath)
	target.page = firstPage - 1

	go target.writeResult(ctx)
	return &target, nil
}

func (s *session) fileName(url *url.URL) string {
	s.sessionMu.Lock()
	s.fileNum += 1
	s.sessionMu.Unlock()

	return fmt.Sprintf("%s/%s_%d.htm", s.dataPath, url.Path, s.fileNum)
}

func (s *session) currentPage() int {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.page += 1
	return s.page
}

func (s *session) appendItem(item Item) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.items = append(s.items, item)
}

func (s *session) writeResult(ctx context.Context) {
	resultFileName := s.dataPath + "/" + "result.csv"
	metadataFileName := s.dataPath + "/" + "header.csv"

	f, err := os.OpenFile(resultFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Panicf("Can't create result file: %v", err)
	}
	defer f.Close()
	wr := csv.NewWriter(f)

	mf, err := os.OpenFile(metadataFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Panicf("Can't create header file: %v", err)
	}
	defer mf.Close()
	mwr := csv.NewWriter(mf)
	err = mwr.Write([]string{"ProductID", "Price", "Title", "HREF", "Desc", "GameTime", "NumberOfPlayers", "Age", "SrcPageRef"})
	if err != nil {
		log.Panicf("Can't write metadata: %v", err)
	}
	mwr.Flush()

	flushFn := func() {
		if len(s.items) == 0 {
			return
		}
		s.sessionMu.Lock()
		defer s.sessionMu.Unlock()
		buf := s.items
		s.items = []Item{}
		for _, i := range buf {
			err := wr.Write([]string{i.ProductID, i.Price, i.Title, i.HREF, i.Desc, i.GameTime, i.NumberOfPlayers, i.Age, i.SrcPageRef})
			if err != nil {
				wr.Flush()
				log.Panicf("Can't write results: %v", err)
			}
		}
		wr.Flush()
		log.Printf("flush data: %d records \n", len(buf))
	}

	for {
		select {
		case <-ctx.Done():
			flushFn()
			log.Print("Function writeResult stopped due to canceled context")
			return
		default:
			flushFn()
			time.Sleep(flushInterval)
		}
	}
}
