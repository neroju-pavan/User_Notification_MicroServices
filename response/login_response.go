package response

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    UserID       int64  `json:"user_id"`
}
