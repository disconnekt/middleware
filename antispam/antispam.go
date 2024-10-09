package antispam

import (
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const BlockTime = 10

type counters struct {
	count      uint64
	expiration uint64
}

type middleware struct {
	list map[string]*counters
	mu   sync.Mutex
}

var instance *middleware

func getInstance() *middleware {
	if instance == nil {
		instance = &middleware{list: make(map[string]*counters)}

		go func() {
			ticker := time.NewTicker(BlockTime * time.Second)
			defer ticker.Stop()

			for {
				<-ticker.C
				instance.mu.Lock()
				now := uint64(time.Now().Unix())
				for ip, c := range instance.list {
					if c.expiration < now {
						delete(instance.list, ip)
					}
				}
				instance.mu.Unlock()
			}
		}()
	}

	return instance
}

func (m *middleware) validRequest(ip string) bool {
	m.mu.Lock()
	counter, ok := m.list[ip]
	if !ok {
		counter = &counters{
			count:      0,
			expiration: uint64(time.Now().Unix()) + BlockTime,
		}
		m.list[ip] = counter
	}
	counter.increment()
	m.mu.Unlock()

	if counter.count > 3 {
		log.Printf("Blocking IP: %s, count: %d, Block Expires: %s \n",
			ip,
			counter.count,
			time.Unix(int64(counter.expiration), 0).UTC().Format(time.RFC3339))

		return false
	}

	return true
}

func (c *counters) increment() {
	if c.count < math.MaxUint64 {
		c.count++
	}
	if c.expiration <= math.MaxUint64-BlockTime {
		c.expiration += BlockTime
	}
}

func WrapFiber(next fiber.Handler, blockFunc fiber.Handler) fiber.Handler {
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

		if getInstance().validRequest(ip) {
			return next(c)
		}

		return blockFunc(c)
	}
}
