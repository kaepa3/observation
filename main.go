package main

import (
	"fmt"
	"time"

	"context"

	"github.com/BurntSushi/toml"
	"github.com/kaepa3/sbth"
)

var config Config

type Config struct {
	Address string
}

func main() {
	readConfig()

	streamThermo := getTemperture()
	streamPic := takePicture()

	text := <-streamThermo
	s := <-streamPic

	tweet(text, s)
}
func readConfig() {
	toml.DecodeFile("config.toml", &config)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f 湿度：%d 電池：%d\n", th.GetTemperature(), th.GetHumidity(), th.GetBattery())
}

func getTemperture() <-chan string {
	ctx, _ := context.WithCancel(context.Background())
	valStream := make(chan string)
	timer := time.NewTimer(time.Second * 8)
	go func() {
		defer close(valStream)
		ch := sbth.Scan(config.Address, ctx)
		select {
		case p := <-ch:
			valStream <- createTweetText(p)
			break
		case <-ctx.Done():
		case <-timer.C:
			valStream <- "Thermohygro Error"
			break
		}
	}()
	return valStream
}
func takePicture() <-chan string {
	valStream := make(chan string)
	go func() {
		defer close(valStream)
		time.Sleep(time.Second * 1)
		valStream <- "pic"
	}()
	return valStream
}
func tweet(text string, imgPath string) {
	fmt.Println(text)
	fmt.Println(imgPath)
}
