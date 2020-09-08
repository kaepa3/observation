package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
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
	Text        string
}

func main() {
	readConfig()

	streamText := getTemperture()

	streamPic := takePicture()
	text := <-streamText
	s := <-streamPic
	tweet(text, s)
}
func readConfig() {
	toml.DecodeFile("config.toml", &conf)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f 湿度：%d 電池：%d\n", th.GetTemperature(), th.GetHumidity(), th.GetBattery(), conf.Text)
}

func getTemperture() <-chan string {
	ctx, _ := context.WithCancel(context.Background())
	valStream := make(chan string)
	log.Println("timer:" + strconv.Itoa(conf.Timeout))
	timer := time.NewTimer(time.Second * time.Duration(conf.Timeout))
	log.Println("search:" + conf.Address)
	ch := sbth.Scan(conf.Address, ctx)

	go func() {
		defer close(valStream)
		select {
		case p := <-ch:
			log.Println("come!!!!")
			valStream <- createTweetText(p)
			break
		case <-ctx.Done():
			log.Println("Done!!!!")
			valStream <- "Thermohygro Error"
			break
		case <-timer.C:

			log.Println("time!!!!")
			valStream <- "Timeout Error"
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
			log.Println(err)
		}
		log.Println("take picture")
		exec.Command("sudo", "raspistill", "-rot", "180", "-o", file).Run()
		log.Println("take finish")
		valStream <- file
	}()
	return valStream
}
func tweet(text string, imgPath string) {

	log.Println("tweet start")
	api := tweetapi.GetTwitterApi(conf.TwitterConf)
	api.Tweet(text, imgPath)
	log.Println("tweet end")
}
