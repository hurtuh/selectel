package transport

import (
	"net/http"
	"strings"
	"time"
)

type lastDuration struct {
	time  int
	count int
}

var ld lastDuration

type CustomRoundTripper struct {
	original   http.RoundTripper
	limit      int
	accounting time.Duration
	exception  []string
	returnFlag bool
}

func (c CustomRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if !searchExceptions(c.exception, r.URL.String()) {
		return c.original.RoundTrip(r)
	}
	if checkDuration(int(c.accounting.Minutes()), c.limit) {
		return c.original.RoundTrip(r)
	} else if !c.returnFlag {
		return c.RoundTrip(r)
	}
	return nil, nil
}

func checkDuration(t int, limit int) bool {
	if t == 0 {
		return true
	}
	if ld.time == 0 {
		ld.time = time.Now().Minute() + t
		return true
	}
	if t+time.Now().Minute()-ld.time <= 0 {
		ld.time = time.Now().Minute() + t
		ld.count = 0
		return true
	}
	if ld.count < limit {
		ld.count++
		return true
	}
	return false
}

func searchExceptions(exception []string, url string) bool {
	for _, v := range exception {
		if strings.Contains(v, "*") {
			index := strings.Index(v, "*")
			if strings.HasPrefix(url, v[0:index-1]) && strings.HasSuffix(url, v[index+1:]) {
				return true
			}
		}
		if v == url {
			return true
		}
	}
	return false
}

func NewThrottler(transport http.RoundTripper, limit int, dur time.Duration, exception []string, rf bool) CustomRoundTripper {
	return CustomRoundTripper{
		original:   transport,
		limit:      limit,
		accounting: dur,
		exception:  exception,
		returnFlag: rf,
	}
}
