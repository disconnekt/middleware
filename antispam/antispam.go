package antispam

import (
	"log"
	"net/http"
	"time"
)

type counters struct {
	Count uint64
	Exp   uint64
}

func (c *counters) increment() {
	if c.Count < 18446744073709551615 {
		c.Count++
	}
	if c.Exp < 18446744073709551585 {
		c.Exp += 30
	}
}

func Wrap(next http.Handler, blockFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	var requestList map[string]*counters

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nil == requestList {
			requestList = make(map[string]*counters)
		}

		ip := r.Header.Get("X-Real-Ip")

		if _, ok := requestList[ip]; !ok {
			requestList[ip] = &counters{
				0,
				uint64(time.Now().Unix()),
			}
			go func() {
				for {
					time.Sleep(30 * time.Second)
					if requestList[ip].Exp < uint64(time.Now().Unix()) {
						delete(requestList, ip)
						break
					}
				}
			}()
		}

		requestList[ip].increment()

		if requestList[ip].Count > 3 {
			log.Println("Block "+ip+" ", requestList[ip].Count)
			blockFunc(w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}
