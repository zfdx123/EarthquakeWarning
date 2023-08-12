package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type No struct {
	Type      string `json:"type"`
	Time      string `json:"time"`
	Location  string `json:"location"`
	Magnitude string `json:"magnitude"`
	Depth     string `json:"depth"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type config struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	EnableMail  bool   `json:"enableMail"`
	AuthCode    string `json:"authCode"`
	SandMali    string `json:"sandMali"`
	SendName    string `json:"sendName"`
	ServiceHost string `json:"serviceHost"`
	ServicePort int    `json:"servicePort"`
	Receive     string `json:"receive"`
}

//var QuakeT map[string]interface{}

func main() {
	var info config
	open, err := os.Open("./config/config.json")
	if err != nil {
		log.Println("配置文件错误：", err.Error())
		return
	}
	defer func(open *os.File) {
		err := open.Close()
		if err != nil {
			log.Println("Close:", err.Error())
		}
	}(open)

	all, err := io.ReadAll(open)
	if err != nil {
		log.Println("读取配置文件错误：", err.Error())
		return
	}

	err = json.Unmarshal(all, &info)
	if err != nil {
		log.Println("配置文件格式错误:", err.Error())
		return
	}

	i := 0
	for {
		fmt.Println("第", i, "次调用")
		newApi(info)
		time.Sleep(time.Second * 2)
		i++
	}
}

func newApi(info config) {
	var (
		client *http.Client
		wg     sync.WaitGroup
		maps   map[string]interface{}
		mapM   map[string]*No
		api    = "https://api.wolfx.jp/cenc_eqlist.json"
	)
	maps = make(map[string]interface{})
	mapM = make(map[string]*No)

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}

	newRequest, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		log.Println("NewRequest:", err.Error())
		return
	}
	do, err := client.Do(newRequest)
	if err != nil {
		log.Println("Do:", err.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Close:", err.Error())
		}
	}(do.Body)

	allBody, err := io.ReadAll(do.Body)
	if err != nil {
		log.Println("ReadAll Body:", err.Error())
		return
	}

	err = json.Unmarshal(allBody, &maps)
	if err != nil {
		log.Println("Json Body:", err.Error())
		return
	}
	delete(maps, "md5")
	marshal, err := json.Marshal(maps)
	if err != nil {
		log.Println("Json Marshal:", err.Error())
		return
	}
	err = json.Unmarshal(marshal, &mapM)
	if err != nil {
		log.Println("Json Body:", err.Error())
		return
	}

	No0 := mapM["No1"]

	// Test Data
	//No0.Type = "automatic"
	//No0.Location = "山东德州市平原县"
	//No0.Depth = "10"
	//No0.Latitude = "39.16"
	//No0.Longitude = "126.34"
	//No0.Magnitude = "6.4"

	qLatitude, err := strconv.ParseFloat(No0.Latitude, 64)
	if err != nil {
		log.Println("ParseFloat:", err.Error())
		return
	}
	qLongitude, err := strconv.ParseFloat(No0.Longitude, 64)
	if err != nil {
		log.Println("ParseFloat:", err.Error())
		return
	}
	qMagnitude, err := strconv.ParseFloat(No0.Magnitude, 64)
	if err != nil {
		log.Println("ParseFloat:", err.Error())
		return
	}
	qDepth, err := strconv.ParseFloat(No0.Depth, 64)
	if err != nil {
		log.Println("ParseFloat:", err.Error())
		return
	}

	dis := EarthDistance(qLatitude, qLongitude, info.Latitude, info.Longitude)

	earthquakeIntensity := calculateEarthquakeIntensity(qMagnitude, qDepth)
	intensityAtDistance := calculateIntensityAtDistance(earthquakeIntensity, dis)

	if info.EnableMail {
		wg.Add(3)
	} else {
		wg.Add(2)
	}

	// No0.Type == "automatic" && strings.Contains(No0.Location, "山东") || No0.Type == "automatic" && dis <= 500.00
	if No0.Type == "automatic" && intensityAtDistance > 2.5 {
		travelTime := calculateTravelTime(dis, 7.0)
		go playMusic(&wg)
		go countdown(travelTime, &wg, intensityAtDistance)
		msg := fmt.Sprintf("当前位置预测强度: %f, 距离: %f, 预计到达时间: %s, 地点: %s, 时间: %s, 震级: %s, 纬度: %s, 经度: %s, 深度: %s", intensityAtDistance, dis, travelTime, No0.Location, No0.Time, No0.Magnitude, No0.Latitude, No0.Longitude, No0.Depth)
		if info.EnableMail {
			go SendMail(info.SandMali, info.AuthCode, info.Receive, info.SendName, info.ServiceHost, info.ServicePort, "地震预警", msg, &wg)
		}
		fmt.Println("****************************************************************************************************************************************************")
		fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 40, 31, msg, 0x1B)
		fmt.Println("****************************************************************************************************************************************************")
	} else {
		if info.EnableMail {
			wg.Add(-3)
		} else {
			wg.Add(-2)
		}
		idx := []string{"No1", "No2", "No3", "No4"}
		fmt.Println("###################################################################################################")
		for _, i := range idx {
			Nox := mapM[i]
			msg := fmt.Sprintf("地点: %s, 时间: %s, 震级: %s, 纬度: %s, 经度: %s, 深度: %s", Nox.Location, Nox.Time, Nox.Magnitude, Nox.Latitude, Nox.Longitude, Nox.Depth)
			fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 40, 32, msg, 0x1B)
		}
		fmt.Println("###################################################################################################")
	}
	wg.Wait()
}

func EarthDistance(latitude1, longitude1, latitude2, longitude2 float64) float64 {
	radLat1 := math.Pi * latitude1 / 180
	radLat2 := math.Pi * latitude2 / 180

	theta := longitude1 - longitude2
	radTheta := math.Pi * theta / 180

	dist := math.Sin(radLat1)*math.Sin(radLat2) + math.Cos(radLat1)*math.Cos(radLat2)*math.Cos(radTheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344
	return dist
}

// 计算地震震级
func calculateEarthquakeIntensity(magnitude, depth float64) float64 {
	return magnitude + 2*math.Log10(depth) - 1.5 // 震级-震源深度公式
}

// 计算在某个距离上的强度
func calculateIntensityAtDistance(intensity, distance float64) float64 {
	return intensity - 1.5*math.Log10(distance) // 震级-距离公式
}

func calculateTravelTime(distance float64, speed float64) time.Duration {
	// 计算地震波所需时间
	travelTime := distance / speed
	// 将时间转换为Duration类型
	duration := time.Duration(travelTime) * time.Second
	return duration
}

func countdown(travelTime time.Duration, w *sync.WaitGroup, magnitude float64) {
	defer w.Done()
	for remaining := travelTime; remaining > 0; remaining -= time.Second {
		fmt.Printf("%c[%d;%d;%dm%s%s%s%f%c[0m\n", 0x1B, 0, 40, 31, "地震波剩余到达时间: ", remaining, " 预测强度: ", magnitude, 0x1B)
		time.Sleep(time.Second)
	}
}

func playMusic(w *sync.WaitGroup) {
	defer w.Done()

	f, err := os.Open("./config/war.mp3")
	if err != nil {
		log.Println("无法打开音频文件：", err.Error())
		return
	}
	s, format, err := mp3.Decode(f)
	if err != nil {
		log.Println("解码错误：", err.Error())
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Println("播放器：", err.Error())
	}

	for i := 0; i <= 2; i++ {
		done := make(chan struct{})
		speaker.Play(beep.Seq(s, beep.Callback(func() {
			close(done)
		})))
		<-done
	}
}

func SendMail(sandMali, authCode, mailTo, sendName, host string, port int, subject, body string, w *sync.WaitGroup) {
	defer w.Done()
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(sandMali, sendName))
	m.SetHeader("To", mailTo)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(host, port, sandMali, authCode)
	err := d.DialAndSend(m)
	if err != nil {
		log.Println("DialAndSend: ", err.Error())
	}
}
