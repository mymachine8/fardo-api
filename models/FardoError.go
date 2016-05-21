package models


type FardoError struct {
	Code int `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Errors []DetailedError `json:"errors,omitempty"`
}

type DetailedError struct {
	Domain string `json:"domain,omitempty"`
	Reason string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}
