package dao

type TwitchUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
}

type ResponseUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"displayName"`
	ProfileImageURL string `json:"profileImageUrl"`
	Email           string `json:"email"`
}

type CheatDetectRequest struct {
	Username string `json:"username"`
}

type SummonKillerRequest struct {
	Name string `json:"name"`
}
