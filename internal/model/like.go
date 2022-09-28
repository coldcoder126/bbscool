package model

//点赞 不可撤销

// LikePost 帖子点赞
type LikePost struct {
	UserId int64
	PostId int64
	BaseModel
}

// LikeComment 评论点赞
type LikeComment struct {
	UserId    int64
	CommentId uint64
	BaseModel
}
