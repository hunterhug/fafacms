package model

type Log struct {
	Id           int    `json:"id" xorm:"bigint pk autoincr"`
	Cid          string `json:"cid"`
	Ip           string `json:"ip"`
	Url          string `json:"url"`
	LogTime      int64  `json:"log_time"`
	Ua           string `json:"ua"`
	UserId       int    `json:"user_id" xorm:"bigint index"`
	Flag         bool   `json:"flag"`
	In           string `json:"in" xorm:"TEXT"`
	Out          string `json:"out" xorm:"TEXT"`
	ErrorId      string `json:"error_id"`
	ErrorMessage string `json:"error_message"`
}
