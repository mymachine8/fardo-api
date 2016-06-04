package models


type FardoError struct {
	message string;
}

func (f FardoError) Error() string {
	return f.message;
}