package nporadio2

import (
	"github.com/sirupsen/logrus"
)

type Client struct {
	NowPlaying *NowPlayingService
	Top2000 *Top2000Service
}

func (c *Client) AddLogger(log *logrus.Logger){
	c.Top2000.addLogger(log)
	c.NowPlaying.addLogger(log)
}

func New() *Client {

	client := Client{
		NowPlaying: newNowPlayingService(),
		Top2000: newTop2000Service(),

	}
	client.Top2000.addClient(&client)
	return &client
}