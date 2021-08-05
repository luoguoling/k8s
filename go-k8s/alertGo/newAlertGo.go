package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"log"
	"github.com/gin-gonic/gin"
)

const (
	webHook_Alert = "https://oapi.dingtalk.com/robot/send?access_token=724402cd85e7e80aa5bbbb7d7a89c74db6a3a8bd8fac4c91923ed3f906664ba4"
)
type Message struct {
	MsgType string `json:"msgtype"`
	Text struct {
		Content string `json:"content"`
		Mentioned_list string `json:"mentioned_list"`
		Mentioned_mobile_list string `json:"mentioned_mobile_list"`
	} `json:"text"`

}
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:annotations`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}
//通知消息结构体
type Notification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:receiver`
	GroupLabels       map[string]string `json:groupLabels`
	CommonLabels      map[string]string `json:commonLabels`
	CommonAnnotations map[string]string `json:commonAnnotations`
	ExternalURL       string            `json:externalURL`
	Alerts            []Alert           `json:alerts`
}
//md的结构
type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

type DingTalkMarkdown struct {
	MsgType  string    `json:"msgtype"`
	At       *At       `json:at`
	Markdown *Markdown `json:"markdown"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
//将notication转换为md
// TransformToMarkdown transform alertmanager notification to dingtalk markdow message
func TransformToMarkdown(notification Notification) (markdown DingTalkMarkdown, err error) {
	var Alarm_level string
	switch notification.CommonLabels["severity"] {
	case "critical":
		Alarm_level = "严重"
	case "warning":
		Alarm_level = "警报"
	case "info":
		Alarm_level = "信息"
	}
	//groupKey := notification.GroupLabels
	status := notification.Status

	//annotations := notification.CommonAnnotations

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf(" **当前状态:**%s \n\n",status))
	buffer.WriteString(fmt.Sprintf("**出现的问题:**%s \n\n",notification.CommonLabels["alertname"]))
	buffer.WriteString(fmt.Sprintf("**集群名称:**%s \n\n",notification.CommonLabels["cluster_name"]))
	buffer.WriteString(fmt.Sprintf("**指标收集方式:** %s \n\n",notification.CommonLabels["alert_type"]))
	buffer.WriteString(fmt.Sprintf("**报警级别:**%s \n\n",Alarm_level))

	for _, alert := range notification.Alerts {
		buffer.WriteString(fmt.Sprint(strings.Repeat("-",10)))
		buffer.WriteString(fmt.Sprintf("\n\n"))
		buffer.WriteString(fmt.Sprintf("**警报线:**  %v \n\n",alert.Labels["threshold_value"]))

		//buffer.WriteString(fmt.Sprintf("\n\n **属性:**   %s \n\n",alert.Labels["instance"]))
		//针对node报警会有node信息
		//针对pod报警会有pod信息
		//如果有容器这个标签那么根据容器的标签写入  如果没有按照node的标签来写入
		if alert.Labels["container"] != "" {
			buffer.WriteString(fmt.Sprintf("**容器:**   %s \n\n",alert.Labels["container"]))
			buffer.WriteString(fmt.Sprintf("**容器名称:** %s \n\n",alert.Labels["pod_name"]))
			buffer.WriteString(fmt.Sprintf("**命名空间:**   %s \n\n",alert.Labels["namespace"]))
			buffer.WriteString(fmt.Sprintf("**容器所在节点:**   %s \n\n",alert.Labels["node"]))
		}else {
			buffer.WriteString(fmt.Sprintf("**节点:**   %s \n\n",alert.Labels["host_ip"]))
		}



		//共用
		buffer.WriteString(fmt.Sprintf(" **开始时间:**%s\n\n", alert.StartsAt.Add(8*time.Hour).Format("2006-01-02 15:04:05")))
		buffer.WriteString(fmt.Sprintf("**当前值:**   %s \n\n",alert.Annotations["current_value"]))
	}


	markdown = DingTalkMarkdown{
		MsgType: "markdown",
		Markdown: &Markdown{
			Title: fmt.Sprintf("####alter名称 %s\n\n", notification.CommonLabels["alertname"]),
			Text:  buffer.String(),
		},
		At: &At{
			IsAtAll: false,
		},
	}
	return
}
//获取报警信息
func getAlertInfo(notification Notification) string {
	fmt.Println("getAlertInfo获取的信息是....")
	fmt.Println(notification)
	var m Message
	m.MsgType = "text"
	msg,err := json.Marshal(notification.GroupLabels)
	if err != nil {
		log.Println("notification.GroupLabels Marshal failed",err)
	}
	msg1,err := json.Marshal(notification.CommonAnnotations["summary"])
	if err != nil {
		log.Println("notification.CommonAnnotations Marshal failed",err)
	}
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("项目告警: %v\n",string(msg)))
	buffer.WriteString(fmt.Sprintf("项目Endpoint: %v\n",string(msg1)))
	buffer.WriteString(fmt.Sprintf("项目告警描述: \"我挂了，快来救我^OO^\"\n"))
	buffer.WriteString(fmt.Sprintf("项目Status:%v\n",notification.Status))
	//恢复消息
	var buffer2 bytes.Buffer
	buffer2.WriteString(fmt.Sprintf("项目告警: %v\n",string(msg)))
	buffer2.WriteString(fmt.Sprintf("项目Endpoint: %v\n",string(msg1)))
	buffer2.WriteString(fmt.Sprintf("项目告警描述: \"哈哈哈，我又回来了！！！\"\n"))
	//buffer2.WriteString(fmt.Sprintf("mentioned_mobile_list: %v\n",msgres["mentioned_mobile_list"]))
	buffer2.WriteString(fmt.Sprintf("项目Status:%v\n",notification.Status))
	if notification.Status == "resolved"{
		m.Text.Content = buffer2.String()
	}else {
		m.Text.Content = buffer.String()
	}
	jsons, err := json.Marshal(m)
	resp := string(jsons)
	return resp
}
//钉钉报警
func SendAlertDingMsg(notification Notification) (err error){
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err")
		}
	}()
	markdown, err := TransformToMarkdown(notification)
	if err != nil {
		fmt.Println("格式转换错误!!!")
	}
	data, err := json.Marshal(markdown)
	if err != nil {
		fmt.Println("json格式转换错误")
	}
	fmt.Println("开始发送报警消息!!!")
	fmt.Println(webHook_Alert)
	//content := `{"msgtype": "text",
	//	"text": {"content": "` + msg + `"}
	//}`

	//创建一个请求
	req, err := http.NewRequest("POST", webHook_Alert, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		fmt.Println("钉钉报警请求异常")
	}
	client := &http.Client{}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
		fmt.Println("顶顶报发送异常!!!")
	}
	defer resp.Body.Close()
	return
}
func AlertInfo(c *gin.Context)  {
	var notification Notification
	fmt.Println("接收到的信息是....")
	fmt.Println(notification)
	err := c.BindJSON(&notification)
	fmt.Printf("%#v",notification)
	if err != nil {
		fmt.Println("绑定信息错误!!!")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}else{
		fmt.Println("绑定信息成功")
	}
	fmt.Println("绑定信息成功!!!")
	msg := getAlertInfo(notification)
	fmt.Println("打印的信息是.....")
	fmt.Println(msg)
	err = SendAlertDingMsg(notification)
	if err != nil {
		fmt.Println("消息发送错误")
	}

}
func main()  {
	t := gin.Default()
	t.POST("/Alert",AlertInfo)
	t.GET("/",func(c *gin.Context){
		c.String(http.StatusOK,"关于alertmanager实现钉钉报警的方法!!!!")
	})
	t.Run(":8089")
}



