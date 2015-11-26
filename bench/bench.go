package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/ngaut/log"
	"github.com/ngaut/tso/client"
)

var serverAddress = flag.String("serveraddr", "localhost:1234", "server address")

const (
	total     = 1000 * 10000
	clientCnt = 10
)

func main() {
	go http.ListenAndServe(":6666", nil)
	var wg sync.WaitGroup
	start := time.Now()
	for x := 0; x < clientCnt; x++ {
		wg.Add(1)
		go func() {
			c := client.NewClient(&client.Conf{ServerAddr: *serverAddress})
			defer wg.Done()
			cnt := total / clientCnt
			prs := make([]*client.PipelineRequest, cnt)
			for i := 0; i < cnt; i++ {
				pr := c.GoGetTimestamp()
				prs[i] = pr
			}

			for i := 0; i < cnt; i++ {
				_, err := prs[i].GetTS()
				if err != nil {
					log.Fatal(err)
				}
			}
		}()
	}
	wg.Wait()

	log.Debugf("Total %d, use %v/s", total, time.Since(start).Seconds())
}
