package dtos

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	FirstName   string `json:"firstName" binding:"required"`
	LastName    string `json:"lastName" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

type SocialLoginRequest struct {
	IDToken   string `json:"idToken" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Name      string `json:"name" binding:"required"`
	GoogleID  string `json:"googleId" binding:"required"`
	PhotoURL  string `json:"photoUrl"`
}
