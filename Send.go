package alisms

type SendReq struct {
	CommonReq
	Action        string `json:"Action"`
	Version       string `json:"Version"`
	RegionId      string `json:"RegionId"`
	PhoneNumbers  string `json:"PhoneNumbers"`
	SignName      string `json:"SignName"`
	TemplateParam string `json:"TemplateParam"`
	TemplateCode  string `json:"TemplateCode"`
	OutId         string `json:"OutId"`
}

func GetSend(phone, templateId, param, signName string, client *AlidayuClient) (*SendReq) {
	return &SendReq{
		CommonReq:     client.getCommonReq(),
		Action:        "SendSms",
		Version:       "2017-05-25",
		RegionId:      "cn-hangzhou",
		PhoneNumbers:  phone,
		SignName:      signName,
		TemplateParam: param,
		TemplateCode:  templateId,
		OutId:         "123",
	}
}
