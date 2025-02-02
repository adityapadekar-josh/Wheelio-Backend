package model

type AccessToken struct {
	AccessToken string `json:"accessToken"`
}

type Token struct {
	Token string `json:"token"`
}

type Email struct {
	Email string `json:"email" validate:"required,email"`
}
