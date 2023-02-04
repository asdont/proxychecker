package handlers

type HTTPError struct {
	Error   string `json:"error"`
	Comment string `json:"comment"`
}
