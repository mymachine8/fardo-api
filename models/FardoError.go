package models


type FardoError struct {
	Message string;
}

func (f FardoError) Error() string {
	return f.Message;
}