package httpproxy

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Stats struct {
	addr        string
	reqCounts   map[string]int32
	failCounts  map[string]int32
	relayCounts map[string]int32

	reqChan   chan *http.Request
	failChan  chan *http.Request
	relayChan chan *http.Request

	handleFun chan func(w http.ResponseWriter, r *http.Request)

	waitChan chan struct{}
	retChan  chan string
}

func NewStats() *Stats {
	s := &Stats{
		reqCounts:   make(map[string]int32),
		failCounts:  make(map[string]int32),
		relayCounts: make(map[string]int32),
		reqChan:     make(chan *http.Request, 100),
		failChan:    make(chan *http.Request, 100),
		relayChan:   make(chan *http.Request, 100),
		waitChan:    make(chan struct{}),
		retChan:     make(chan string),
	}

	return s
}

func (p *Stats) Run() {
	go func() {
		for {
			select {
			case req := <-p.reqChan:
				url := hostport(req)
				p.reqCounts[url]++
			case req := <-p.failChan:
				url := hostport(req)
				p.failCounts[url]++
			case req := <-p.relayChan:
				url := hostport(req)
				p.relayCounts[url]++
			case <-p.waitChan:
				type Log struct {
					url        string
					percent    int32
					reqCount   int32
					failCount  int32
					relayCount int32
				}
				logs := make([]Log, 0, len(p.reqCounts))
				for url, reqCount := range p.reqCounts {
					failCount := p.failCounts[url]
					percent := (failCount * 100) / reqCount
					logs = append(logs, Log{
						url:        url,
						percent:    percent,
						reqCount:   reqCount,
						failCount:  failCount,
						relayCount: p.relayCounts[url],
					})
				}
				sort.Slice(logs, func(i, j int) bool {
					if logs[i].percent > logs[j].percent {
						return true
					} else if logs[i].percent < logs[j].percent {
						return false
					} else {
						return logs[i].reqCount > logs[j].reqCount
					}
				})

				var builder strings.Builder
				for _, v := range logs {
					s := fmt.Sprintf("失败率:%d%% 请求:%d 失败:%d 转发:%d 地址:%s", v.percent, v.reqCount, v.failCount, v.relayCount, v.url)
					builder.WriteString(`<font size="4">`)
					builder.WriteString(s)
					builder.WriteString(`</font><br>`)
				}
				p.retChan <- builder.String()
			}
		}
	}()
}

func (p *Stats) LogReq(req *http.Request) {
	p.reqChan <- req
}

func (p *Stats) LogFail(req *http.Request) {
	p.failChan <- req
}

func (p *Stats) LogRelay(req *http.Request) {
	p.relayChan <- req
}

func (p *Stats) WaitStatsLog() string {
	p.waitChan <- struct{}{}
	s := <-p.retChan
	return s
}
