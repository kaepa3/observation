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

	ctx, _ := context.WithCancel(context.Background())
	streamThermo := getTemperture(ctx)
	streamPic := takePicture()

	t := <-streamThermo
	ctx.Done()
	text := createTweetText(t)

	s := <-streamPic

	tweet(text, s)
}
func readConfig() {
	toml.DecodeFile("config.toml", &config)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f湿度：%d電池：%d\n", th.GetTemperature(), th.GetHumidity(), th.GetBattery())
}

func getTemperture(ctx context.Context) <-chan sbth.ThermohygroPacket {
	fmt.Println(config)
	return sbth.Scan(config.Address, ctx)
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
