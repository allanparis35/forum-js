package handlers

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RegisterResponse struct {
    ID    uint   `json:"id"`
    Email string `json:"email"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Captcha  string `json:"captcha"`
}

type LoginResponse struct {
    Token        string `json:"token"`
    RefreshToken string `json:"refresh_token"`
}


