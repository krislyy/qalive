package rtmp

import (
	"fmt"
	"time"
	"github.com/zhangpeihao/goflv"
	rtmp "github.com/zhangpeihao/gortmp"
	"github.com/krislyy/qalive/configure"
	"context"
)

const (
	programName = "RtmpPublisher"
	version     = "0.0.1"
)

type OutboundConnHandler struct {
}

var obConn rtmp.OutboundConn
var createStreamChan chan rtmp.OutboundStream
var videoDataSize int64
var audioDataSize int64
var config *configure.Configure
var cancel context.CancelFunc

var status uint

func (handler *OutboundConnHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	if obConn == nil {
		return
	}
	status, err = obConn.Status()
	fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", status, err)
}

func (handler *OutboundConnHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")
	cancel()
}

func (handler *OutboundConnHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
}

func (handler *OutboundConnHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedRtmpCommand: %+v\n", command)
}

func (handler *OutboundConnHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	createStreamChan <- stream
}
func (handler *OutboundConnHandler) OnPlayStart(stream rtmp.OutboundStream) {

}
func (handler *OutboundConnHandler) OnPublishStart(stream rtmp.OutboundStream) {
	// Set chunk buffer size
	go publish(stream)
}

func publish(stream rtmp.OutboundStream) {
	var err error
	cacheTags := NewCacheTags(20, *config)
	go cacheTags.StartCache()
	fmt.Println("2")
	startTs := uint32(0)
	startAt := time.Now().UnixNano()
	fmt.Println("3")
	for tag := range cacheTags.TagCh {
		if status != rtmp.OUTBOUND_CONN_STATUS_CREATE_STREAM_OK {
			fmt.Println("@@@@@@@@@@@@@@Create stream not ready")
			cacheTags.Quit = true
			break
		}
		if tag.IsFinished {
			fmt.Println("@@@@@@@@@@@@@@File finished")
			startAt = time.Now().UnixNano()
			startTs = uint32(0)
			continue
		}

		switch tag.TagType {
		case flv.VIDEO_TAG:
			videoDataSize += int64(len(tag.Data))
		case flv.AUDIO_TAG:
			audioDataSize += int64(len(tag.Data))
		}

		if startTs == uint32(0) {
			startTs = tag.Timestamp
		}
		diff1 := uint32(0)
		//deltaTs := uint32(0)
		if tag.Timestamp > startTs {
			diff1 = tag.Timestamp - startTs
		} else {
			//fmt.Printf("@@@@@@@@@@@@@@diff1 header(%+v), startTs: %d\n", header, startTs)
		}
		//fmt.Printf("@@@@@@@@@@@@@@diff1 header(%+v), startTs: %d\n", tag, startTs)
		if err = stream.PublishData(tag.TagType, tag.Data, diff1); err != nil {
			fmt.Println("PublishData() error:", err)
			break
		}
		diff2 := uint32((time.Now().UnixNano() - startAt) / 1000000)
		//fmt.Printf("diff1: %d, diff2: %d\n", diff1, diff2)
		if diff1 > diff2+100 {
			//fmt.Printf("header.Timestamp: %d, now: %d\n", header.Timestamp, time.Now().UnixNano())
			time.Sleep(time.Millisecond * time.Duration(diff1-diff2))
		}
	}
}

func RTMP_Publish(conf *configure.Configure)  {
	config = conf
	defer func() {
		err := recover()
		if err != nil {
		}
	}()
	
	createStreamChan = make(chan rtmp.OutboundStream)
	outHandler := &OutboundConnHandler{}
	fmt.Println("to dial")
	fmt.Println("a")
	var err error
	obConn, err = rtmp.Dial(config.Crtmp_url, outHandler, 100)
	if err != nil {
		panic(fmt.Sprintf("Dial error %s", err.Error()))
	}
	fmt.Println("b")
	defer obConn.Close()
	fmt.Println("to connect")
	err = obConn.Connect()
	if err != nil {
		panic(fmt.Sprintf("Connect error: %s", err.Error()))
	}
	fmt.Println("c")
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	for {
		select {
		case stream := <-createStreamChan:
			// Publish
			stream.Attach(outHandler)
			err = stream.Publish(config.StreamName, "live")
			if err != nil {
				str := fmt.Sprintf("Publish error: %s", err.Error())
				panic(str)
			}

		case <-time.After(1 * time.Second):
			fmt.Printf("Audio size: %d bytes; Vedio size: %d bytes\n", audioDataSize, videoDataSize)

		case <-ctx.Done():
			fmt.Println("RTMP stream closed!")
			return
		}
	}
}

func RTMP_Stop(conf *configure.Configure) {
	cancel()
}
