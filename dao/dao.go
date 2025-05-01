package dao

type TwitchUser struct {
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

type ResponseTwitchUser struct {
	Login           string `json:"login"`
	DisplayName     string `json:"displayName"`
	ProfileImageURL string `json:"profileImageUrl"`
}

type AdminTwitchUser struct {
	Login string `json:"login"`
}

type CheatDetectRequest struct {
	Username string `json:"username"`
}

type SummonKillerRequest struct {
	Name string `json:"name"`
}

type AdminLoginResponse struct {
	Token string             `json:"token"`
	User  ResponseTwitchUser `json:"user"`
}
