package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"test123/service"
	"test123/utils"
	"time"
)

func RequirePermission(s *service.AuthorizeService, permission string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID int
			switch v := r.Context().Value(UserIDKey).(type) {
			case string:
				id, err := strconv.Atoi(v)
				if err != nil {
					utils.RespondJSON(w, 500, map[string]string{"error": "invalid user id"})
					return
				}
				userID = id
			case int:
				userID = v
			default:
				utils.RespondJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			cacheKey := "user_perm_" + strconv.Itoa(userID)
			var perms []string

			cached, err := s.Cache.Get(r.Context(), cacheKey).Result()
			if err == nil {
				if uErr := json.Unmarshal([]byte(cached), &perms); uErr != nil {
					log.Println("failed to unmarshal cached permissions:", uErr)
					perms = nil
				}
			}

			//  Fetch from DB if cache miss
			if len(perms) == 0 {
				perms, err = s.GetPermissions(r.Context(), userID)
				if err != nil {
					log.Println("error retrieving permissions from DB:", err)
					utils.RespondJSON(w, 500, map[string]string{"error": "internal server error"})
					return
				}

				//  Populate cache
				data, _ := json.Marshal(perms)
				_ = s.Cache.Set(r.Context(), cacheKey, data, 100*time.Second).Err()
			}

			// Check if required permission exists
			hasPermission := false
			for _, p := range perms {
				if p == permission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				utils.RespondJSON(w, 403, map[string]string{"error": "forbidden - insufficient permissions"})
				return
			}

			//  Permission granted, continue
			next.ServeHTTP(w, r)
		})
	}
}
