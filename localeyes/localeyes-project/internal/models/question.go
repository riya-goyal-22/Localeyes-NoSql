package models

type Question struct {
	QId    string `json:"question_id" dynamodbav:"sk"`
	PostId string `json:"post_id" dynamodbav:"pk"`
	UserId string `json:"q_user_id" dynamodbav:"q_user_id"`
	Text   string `json:"text" dynamodbav:"text"`
}

type Reply struct {
	RId    string `json:"r_id" dynamodbav:"sk"`
	QId    string `json:"q_id" dynamodbav:"pk"`
	Answer string `json:"answer" dynamodbav:"answer"`
	UserId string `json:"r_user_id" dynamodbav:"r_user_id"`
}
