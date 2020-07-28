package main

import (
	"fmt"
	"os/exec"
	"time"

	"context"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/kaepa3/sbth"
	"github.com/kaepa3/tweet/config"
	"github.com/kaepa3/tweet/tweetapi"
)

var conf AppConfig

type AppConfig struct {
	Address     string
	TwitterConf config.TwitterConfig
	Timeout     int
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
	toml.DecodeFile("config.toml", &conf)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f 湿度：%d 電池：%d\n", th.GetTemperature(), th.GetHumidity(), th.GetBattery())
}

func getTemperture() <-chan string {
	ctx, _ := context.WithCancel(context.Background())
	valStream := make(chan string)
	timer := time.NewTimer(time.Second * conf.Timeout)
	go func() {
		defer close(valStream)
		ch := sbth.Scan(conf.Address, ctx)
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
		file := "image.jpg"
		if err := os.Remove(file); err != nil {
			fmt.Println(err)
		}
		exec.Command("sudo", "raspistill", "-o", file).Run()
		valStream <- file
	}()
	return valStream
}
func tweet(text string, imgPath string) {

	api := tweetapi.GetTwitterApi(conf.TwitterConf)
	api.Tweet(text, imgPath)
}
