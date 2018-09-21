package slack

import (
	"log"
	"os"
	"testing"
)

var(
	testSlack *Slack
	testCfg = "testSlackCfg.json"
)

func TestMain(m *testing.M) {
	var err error
	testSlack, err = LoadSlackConfig(testCfg)
	if err != nil {
		log.Fatalf("Failed to load initial configuration for slack test")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestSendMessage(t *testing.T) {
	testSlack.SendMessage(testSlack.Channel, "This is a unit test")
}
