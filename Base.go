package alisms

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const bASE_URL = "http://dysmsapi.aliyuncs.com/"

var Client *AlidayuClient

type AlidayuClient struct {
	AccessKeyId  string `json:"access_key_id"`
	AccessSecret string `json:"access_secret"`
}

type CommonReq struct {
	SignatureMethod  string `json:"SignatureMethod"`
	SignatureNonce   string `json:"SignatureNonce"`
	AccessKeyId      string `json:"AccessKeyId"`
	SignatureVersion string `json:"SignatureVersion"`
	Timestamp        string `json:"Timestamp"`
	Format           string `json:"Format"`
}

func InitAliDaYu(accessKeyId, accessSecret string) {
	Client = &AlidayuClient{AccessKeyId: accessKeyId, AccessSecret: accessSecret}
}

func (this *AlidayuClient) specialUrlEncode(value string) string {
	value = url.QueryEscape(value)
	value = strings.Replace(value, "+", "%20", -1)
	value = strings.Replace(value, "*", "%2A", -1)
	value = strings.Replace(value, "%7E", "~", -1)
	return value
}

func (this *AlidayuClient) sign(str, accessSecret string) string {
	mac := hmac.New(sha1.New, []byte(accessSecret))
	mac.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (this *AlidayuClient) newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func (this *AlidayuClient) getTimeStr() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func (this *AlidayuClient) makeParam(request interface{}) (str string) {
	jsonStr, _ := json.Marshal(request)
	var param = make(map[string]string, 0)
	var paramKeys = make([]string, 0)
	var signstr = this.AccessKeyId
	json.Unmarshal(jsonStr, &param)
	for k, _ := range param {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	for _, v := range paramKeys {
		if v != "Signature" {
			signstr = signstr + v + param[v]
		}
	}
	for _, v := range paramKeys {
		str = str + "&" + this.specialUrlEncode(v) + "=" + this.specialUrlEncode(param[v])
	}
	return
}

func (this *AlidayuClient) getCommonReq() CommonReq {
	uuid, _ := this.newUUID()
	return CommonReq{
		SignatureMethod:  "HMAC-SHA1",
		SignatureNonce:   uuid,
		AccessKeyId:      this.AccessKeyId,
		SignatureVersion: "1.0",
		Timestamp:        this.getTimeStr(),
		Format:           "JSON",
	}
}

func (this *AlidayuClient) get(requestUrl string) (result []byte, err error) {
	httpclient := &http.Client{
		Timeout: time.Duration(time.Second * 10),
	}
	var resp *http.Response
	resp, err = httpclient.Get(requestUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func (this *AlidayuClient) do(request interface{}) (requestId string, err error) {
	paramStr := this.makeParam(request)
	requestBuilder := "GET" + "&" + this.specialUrlEncode("/") + "&" + this.specialUrlEncode(paramStr[1:])
	getUrl := bASE_URL + "?Signature=" + this.specialUrlEncode(this.sign(requestBuilder, this.AccessSecret+"&")) + paramStr
	var result []byte
	result, err = this.get(getUrl)
	if err != nil {
		return
	} else {
		return this.parseResponse(result)
	}
}

func (this *AlidayuClient) parseResponse(response []byte) (requestId string, err error) {
	responseParam := make(map[string]string, 0)
	err = json.Unmarshal(response, &responseParam)
	if err != nil {
		err = errors.New(string(response))
		return
	} else {
		if responseParam["Code"] != "OK" {
			err = errors.New(string(response))
			return
		} else {
			if responseParam["Code"] != "" {
				requestId = responseParam["Code"]
				return
			} else {
				err = errors.New(string(response))
				return
			}
		}
	}
}

//短信发送 https://help.aliyun.com/document_detail/56189.html?spm=5176.doc55451.6.570.NggOGq
func (this *AlidayuClient) SendSms(phone, tempalteId, content, signName string, param map[string]string) (requestId string, err error) {
	data, _ := json.Marshal(param)
	request := GetSend(phone, tempalteId, string(data), signName, this)
	return this.do(request)
}
