package main

import (
	"fmt"
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
}

func main() {
	readConfig()

	text := getTemperture()
	streamPic := takePicture()
	s := <-streamPic
	tweet(text, s)
}
func readConfig() {
	toml.DecodeFile("config.toml", &conf)
}

func createTweetText(th sbth.ThermohygroPacket) string {
	return fmt.Sprintf("温度：%.2f 湿度：%d 電池：%d\n#枝豆日記", th.GetTemperature(), th.GetHumidity(), th.GetBattery())
}

func getTemperture() string {
	ctx, _ := context.WithCancel(context.Background())
	text := ""
	fmt.Println("timer:" + strconv.Itoa(conf.Timeout))
	timer := time.NewTimer(time.Second * time.Duration(conf.Timeout))
	fmt.Println("search:" + conf.Address)
	ch := sbth.Scan(conf.Address, ctx)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case p := <-ch:
				fmt.Println("come!!!!")
				text = createTweetText(p)
				close(done)
				return
			case <-ctx.Done():
				fmt.Println("Done!!!!")
				text = "Thermohygro Error"
				close(done)
				return
			case <-timer.C:

				fmt.Println("time!!!!")
				text = "Timeout Error"
				close(done)
				return
			}
		}
		fmt.Println("done!!!!")
	}()
	<-done
	return text
}
func takePicture() <-chan string {
	valStream := make(chan string)
	go func() {
		defer close(valStream)
		file := "image.jpg"
		fmt.Println("delete image")
		if err := os.Remove(file); err != nil {
			fmt.Println(err)
		}
		fmt.Println("take picture")
		exec.Command("sudo", "raspistill", "-rot", "90", "-o", file).Run()
		valStream <- file
	}()
	return valStream
}
func tweet(text string, imgPath string) {

	api := tweetapi.GetTwitterApi(conf.TwitterConf)
	api.Tweet(text, imgPath)
}
