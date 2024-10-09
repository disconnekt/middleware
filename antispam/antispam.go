package antispam

import (
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
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

func WrapFiber(next fiber.Handler, blockFunc fiber.Handler) fiber.Handler {
	var (
		mu          sync.Mutex
		requestList = make(map[string]*counters)
	)

	go func() {
		ticker := time.NewTicker(BlockTime * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
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

	return func(c *fiber.Ctx) error {
		ip := c.Get("X-Real-Ip")
		if ip == "" {
			ip = c.Get("X-Forwarded-For")
			if ip != "" {
				ip = strings.Split(ip, ",")[0]
			}
		}
		if ip == "" {
			ip = c.IP()
		}

		mu.Lock()
		counter, ok := requestList[ip]
		if !ok {
			counter = &counters{
				Count:      0,
				Expiration: uint64(time.Now().Unix()) + BlockTime,
			}
			requestList[ip] = counter
		}
		counter.increment()
		mu.Unlock()

		if counter.Count > 3 {
			log.Printf("Blocking IP: %s, Count: %d, BlockTime: %d seconds\n", ip, counter.Count, BlockTime)
			return blockFunc(c)
		}

		return next(c)
	}
}
