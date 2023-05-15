package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/akbarhabiby/go-api-test/helpers"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiter(next http.Handler) http.Handler {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  60,
	}
	limiter := limiter.New(memory.NewStore(), rate)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identifier := helpers.GetRealIP(req)

		l, _ := limiter.Get(req.Context(), identifier)

		h := w.Header()
		h.Set("X-RateLimit-Limit", strconv.FormatInt(l.Limit, 10))
		h.Set("X-RateLimit-Remaining", strconv.FormatInt(l.Remaining, 10))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(l.Reset, 10))

		if l.Reached {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(("???")))
			return
		}

		next.ServeHTTP(w, req)
	})
}
