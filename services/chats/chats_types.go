package chats

type createChatRequest struct {
    Name              string `json:"name"`
	TypeOfChat        string   `json:"type_of_chat"`
	UsersInvolvedInChat []string `json:"users_involved_in_chat"`
}

type CheckChatRequest struct {
	UserID string `json:"user_id"`
}