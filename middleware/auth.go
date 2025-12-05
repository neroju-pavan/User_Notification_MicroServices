package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"test123/service"
	"test123/utils"
	"time"
)

type authContextKey string

const UserIDKey authContextKey = "userID"

func AuthMiddleware(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Authorize request
			userID, err := auth.Authorize(r.Context(), r)
			if err != nil {
				utils.RespondJSON(w, 401, map[string]string{
					"error": "unauthorized",
				})
				return

			}

			//rate limit 100 request in a minute

			key := fmt.Sprintf("rate_limit:user:%d", userID)
			count, _ := auth.Redis.Incr(r.Context(), key).Result()

			if count == 1 {
				auth.Redis.Expire(r.Context(), key, 2*time.Minute)
			}

			if count > 100 {
				utils.RespondJSON(w, 429, map[string]string{
					"error": "rate limit exceeded, try later",
				})
				return
			}

			// Inject userID into request context
			log.Println(userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
