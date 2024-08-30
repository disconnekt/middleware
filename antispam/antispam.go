package antispam

import (
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

const BlockTime = 5

type counters struct {
	Count      uint64
	Expiration uint64
}

func (c *counters) increment() {
	if c.Count < math.MaxUint64 {
		c.Count++
	}
	if c.Expiration <= math.MaxUint64-BlockTime {
		c.Expiration += BlockTime
	}
}

func Wrap(next http.Handler, blockFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	var (
		mu          sync.Mutex
		requestList = make(map[string]*counters)
	)

	go func() {
		for {
			time.Sleep(BlockTime * time.Second)
			mu.Lock()
			now := uint64(time.Now().Unix())
			for ip, c := range requestList {
				if c.Expiration < now {
					delete(requestList, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-Ip")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		mu.Lock()
		if _, ok := requestList[ip]; !ok {
			requestList[ip] = &counters{
				Count:      0,
				Expiration: uint64(time.Now().Unix()) + BlockTime,
			}
		}
		requestList[ip].increment()
		mu.Unlock()

		if requestList[ip].Count > 3 {
			log.Println("Block "+ip+" ", requestList[ip].Count)
			blockFunc(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
