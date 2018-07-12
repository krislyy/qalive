package rtmp

import (
	"testing"

	"gome.com/qalive/configure"
	"sync"
)

func TestPublish(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	config := &configure.Configure{
		DefaultDir: "/Users/kaolafm/Desktop/",
		Crtmp_url:"rtmp://10.112.179.9:1935/live",
		StreamName:"movie",
		PlayList: []string{"demo1", "demo2", "demo3"},
	}
	go RTMP_Publish(config)
	wg.Wait()
}