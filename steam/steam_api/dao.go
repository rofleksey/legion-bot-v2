package steam_api

type CommentsResponse struct {
	Success      bool   `json:"success"`
	CommentsHTML string `json:"comments_html"`
}
