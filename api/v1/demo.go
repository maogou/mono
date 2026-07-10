package v1

type AddAuthRequest struct {
	ParkCode int64 `json:"park_code"`
}

type AddAuthResponse struct {
	ParkCode int64 `json:"park_code"`
}
