package main

import (
	"fmt"
	"os/exec"

	"context"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/kaepa3/sbth"
	"github.com/kaepa3/tweet/tweetapi"
)

var config Config

type Config struct {
	Address     string
	TwitterConf config.TwitterConfig
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
		file := "image.jpg"
		if err := os.Remove(file); err != nil {
			fmt.Println(err)
		}
		err := exec.Command("sudo", "raspistill", "-o", file).Run()
		valStream <- file
	}()
	return valStream
}
func tweet(text string, imgPath string) {

	api := tweetapi.GetTwitterApi(*conf, Tweet)
	api.Tweet(text, imgPath)
}
