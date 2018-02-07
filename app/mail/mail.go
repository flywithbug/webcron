package mail

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils"
	"time"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"net/http"
	"webcron/app/models"

	"strings"
	"strconv"
)

var (
	sendCh chan *utils.Email
	config string
)

func init() {
	queueSize, _ := beego.AppConfig.Int("mail.queue_size")
	host := beego.AppConfig.String("mail.host")
	port, _ := beego.AppConfig.Int("mail.port")
	username := beego.AppConfig.String("mail.user")
	password := beego.AppConfig.String("mail.password")
	from := beego.AppConfig.String("mail.from")
	if port == 0 {
		port = 25
	}

	config = fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","port":%d,"from":"%s"}`, username, password, host, port, from)

	sendCh = make(chan *utils.Email, queueSize)

	go func() {
		for {
			select {
			case m, ok := <-sendCh:
				if !ok {
					return
				}
				if err := m.Send(); err != nil {
					beego.Error("SendMail:", err.Error())
				}
			}
		}
	}()
}


func SendMail(address, name, subject, content string, cc []string) bool {
	fmt.Println(address,name,subject,content,cc)
	mail := utils.NewEMail(config)
	mail.To = []string{address}
	mail.Subject = subject
	mail.HTML = content
	if len(cc) > 0 {
		mail.Cc = cc
	}

	select {
	case sendCh <- mail:
		return true
	case <-time.After(time.Second * 3):
		return false
	}
}

func SendMsg(subject string,content string,cc []string)bool  {
	para :=make(map[string]interface{})
	para["receivers"]= cc
	para["title"] = subject
	para["content"] = content
	para["group_id"]= "mta-crontab-notify"
	var header map[string]string
	header = make(map[string]string)
	header["Content-Type"]="application/json;charset=utf-8"
	_,err :=  POST("http://10.66.3.50:8188/sendmsg",para,header)
	if err != nil{
		beego.Error(err.Error())
		return false
	}
	return true
}
func SendMsgWithTask(task* models.Task )  {
	fmt.Println("task",task.Id)

	ccList := make([]string, 0)
	if task.NotifyEmail != "" {
		ccList = strings.Split(task.NotifyEmail, "\n")
	}
	user,err := models.UserGetById(task.Id)
	if err != nil{
		beego.Error(err.Error())
	}else {
		ccList = append(ccList, user.UserName)
	}
	fmt.Println(ccList)

	url := "http://10.66.3.50:5000/nuwa/newton/api/mta/applastestver?appid="
	b ,err := GET(url+strconv.Itoa(task.Id))
	if err != nil {
		beego.Error(err.Error())
		return
	}
	var result map[string]interface{}
	err = json.Unmarshal(b,&result)
	if err != nil {
		beego.Error(err.Error())
	}
	fmt.Println(result)
	title := task.TaskName + "：定时任务启动"
	content := result["data"]
	fmt.Println(title,content)
	if len(ccList) != 0 {
		SendMsg(title,content.(string),ccList)

	}
}


func POST(url string,v interface{},header map[string]string) ([]byte, error)  {
	j,err := json.Marshal(v)
	fmt.Printf(string(j))
	if err !=nil {
		return nil,err
	}
	req , err := http.NewRequest("POST",url,bytes.NewBuffer(j))
	for k,v := range header  {
		req.Header.Set(k,v)
	}
	client := &http.Client{}
	resp ,err := client.Do(req)
	if err != nil{
		return nil,err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	return body,err
}


func GET(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

