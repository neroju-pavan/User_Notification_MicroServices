package requests

import "test123/errors"

type LoginTokenRequest struct {
	AccessToken  string `json:"access"`
	RefreshToken string `json:"refresh"`
}

func (r *LoginTokenRequest) Validate() error {

	if r.AccessToken == "" || r.RefreshToken == "" {
		return errors.ErrMissingField
	}
	return nil
}