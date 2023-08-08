# 地震预警

# 请注意本项目会尽可能的提前告知用户，但不能保障稳定和是否会出现突发状况，请不要过度依赖该项目

本项目着手来源于山东平原地震，在手机音量的限制下，叫人醒来需要更多的时间，所以接入api电脑实时获取数据，以便获得更多的逃生时间

## 感谢wolfx提供的API和站长素材提供的音频文件

> https://api.wolfx.jp/
>
> https://sc.chinaz.com/


## 配置文件
请在运行前修改配置文件 config/config.json


```
{
  "latitude": xx.xxxxx,
  "longitude": xxx.xxxxx,
  "enableMail": true, # 是否启用地震邮箱发送
  "authCode": "XXXXX", # 授权码
  "sandMali": "xx@x.x", # 发件人邮箱
  "sendName": "xxx", # 发件人用户名
  "serviceHost": "smtp.163.com", # 发件服务器SSL smtp
  "servicePort": 465, # 服务器端口
  "receive": "xx@x.x" # 收件人邮箱
}

```


