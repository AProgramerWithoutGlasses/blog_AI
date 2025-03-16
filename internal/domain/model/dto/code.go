package dto

type CodeReq struct {
	Question string
	UserId   uint
	CodeType string
}

type Code struct {
	ID          uint
	Key         string
	Question    string
	Explanation string
}
