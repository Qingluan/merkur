package merkur

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Qingluan/merkur/config"
)

const (
	Random = 1
	Flow   = 0
)

var (
	OrderHistory = map[string]*ProxyPool{}
	lock         sync.Mutex
)

type Result struct {
	uri string
	res interface{}
}

type ProgressBar interface {
	Add(i int) int
	Increment() int
	Update()
	// SetTotal(all int) *ProgressBar
}

type ProxyPool struct {
	res          map[string]int
	Mode         int
	now          int
	threadnum    int
	lastDownload time.Time
}

// NewProxyPool can set mode pool.Mode = Random / pool.Mode = Flow (default is flow)
func NewProxyPool(url ...string) (proxyPool *ProxyPool) {
	if url != nil {
		proxyPool = &ProxyPool{
			res:          make(map[string]int),
			lastDownload: time.Now(),
		}
		switch url[0][:5] {
		case "ss://":
			proxyPool.res[url[0]] = 0
		case "ssr:/":
			proxyPool.res[url[0]] = 0
		case "vmess":
			proxyPool.res[url[0]] = 0
		case "socks":
			proxyPool.res[url[0]] = 0
		case "https":
			for no, i := range config.ParseOrding(url[0]) {
				proxyPool.res[i] = no
			}
			lock.Lock()
			defer lock.Unlock()
			OrderHistory[url[0]] = proxyPool
		case "http:":
			for no, i := range config.ParseOrding(url[0]) {
				proxyPool.res[i] = no
			}
			lock.Lock()
			defer lock.Unlock()
			OrderHistory[url[0]] = proxyPool

		default:
			log.Println("agly url :", url[0])
		}
	} else {
		proxyPool = &ProxyPool{
			res: make(map[string]int),
		}
	}

	return
}

func (proxyPool *ProxyPool) Empty() bool {
	if len(proxyPool.res) == 0 {
		return true
	}
	return false
}

func (proxyPool *ProxyPool) Count() int {
	return len(proxyPool.res)
}

func (pool *ProxyPool) LoopOneTurn(doWhat func(proxyDialer Dialer) interface{}, processor ...ProgressBar) (o map[string]interface{}) {
	// ordnum := len(pool.res)
	o = make(map[string]interface{})
	threadnum := pool.threadnum
	if threadnum == 0 {
		threadnum = 30
	}
	wait := sync.WaitGroup{}
	taskRes := make(chan Result, 1024)
	var bar ProgressBar
	if processor != nil {
		bar = processor[0]
		// bar.SetTotal(ordnum)
		// fmt.Println =
	}

	go func() {
		for one := range taskRes {
			if one.uri == "[stop]" {
				break
			}
			if bar != nil {
				bar.Increment()
				bar.Update()
			}
			o[one.uri] = one.res

			// }
		}
	}()

	for k := range pool.res {
		wait.Add(1)
		go func(uri string, p chan Result, bar ProgressBar) {
			defer func() {
				wait.Done()

			}()
			dialer, err := NewDialerByURI(uri)
			if err != nil {
				p <- Result{uri, err}
				return
			}
			res := doWhat(dialer)
			p <- Result{uri, res}
		}(k, taskRes, bar)
	}
	wait.Wait()
	taskRes <- Result{"[stop]", nil}

	return
}

func ArraiExists(arr []string, o string) bool {
	for _, i := range arr {
		if i == o {
			return true
		}
	}
	return false
}

func (proxyPool *ProxyPool) Adds(urls []string) {
	oldNum := len(proxyPool.res)

	for _, i := range urls {
		if _, ok := proxyPool.res[i]; !ok {
			proxyPool.res[i] = oldNum
			oldNum++
		}
	}
}

func (proxyPool *ProxyPool) Urls() (urls []string) {
	for i := range proxyPool.res {
		urls = append(urls, i)
	}
	return
}

func (proxyPool *ProxyPool) Add(url string) {
	oldNum := len(proxyPool.res)
	if strings.HasPrefix(url, "http") {
		if k, ok := OrderHistory[url]; ok && time.Now().Sub(k.lastDownload) < 1*time.Minute {
			proxyPool.Merge(*k)
			// lock.Lock()
			// defer lock.Unlock()
			// proxyPool.lastDownload = k.lastDownload
			// delete(OrderHistory, url)
		} else {
			proxyPool.lastDownload = time.Now()
			lock.Lock()
			defer lock.Unlock()
			OrderHistory[url] = proxyPool
			for _, i := range config.ParseOrding(url) {
				if _, ok := proxyPool.res[i]; !ok {
					proxyPool.res[i] = oldNum
					oldNum++
				}
			}
		}

	} else if strings.HasPrefix(url, "ss") {
		proxyPool.res[url] = oldNum
	} else if strings.HasPrefix(url, "vmess://") {
		proxyPool.res[url] = oldNum
	}
}

func (pool *ProxyPool) Merge(ppool ProxyPool) {
	oldnum := len(pool.res)
	for k := range ppool.res {
		pool.res[k] = oldnum
		oldnum++
	}
}

func (pool *ProxyPool) Get() (outuri string) {
	var u string
	if pool.Mode == Random {
		// i :=
		ii := rand.Int() % len(pool.res)
		for u, i := range pool.res {
			if i == ii {
				return u
			}
		}
	} else {
		for u, i := range pool.res {
			if i == pool.now%len(pool.res) {
				return u
			}
		}
	}
	defer func() {
		pool.now++
		pool.now %= len(pool.res)
	}()
	return u
}
