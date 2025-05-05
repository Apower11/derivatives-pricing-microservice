package followers

type FollowRequest struct {
	FolloweeUserID string `json:"followee_user_id"`
}

type FollowCheckRequest struct {
	FolloweeUserID string `json:"followee_user_id"`
}