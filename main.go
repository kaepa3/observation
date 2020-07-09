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

	ctx, _ := context.WithCancel(context.Background())
	streamThermo := getTemperture(ctx)
	streamPic := takePicture()

	t := <-streamThermo
	text := createTweetText(t)

	s := <-streamPic

	tweet(text, s)
}
func readConfig() {
	toml.DecodeFile("config.toml", &config)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f湿度：%.2f\n", th.GetTemperature(), th.GetHumidity())
}

func getTemperture(ctx context.Context) <-chan sbth.ThermohygroPacket {
	return sbth.Scan(config.Address, ctx)
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
