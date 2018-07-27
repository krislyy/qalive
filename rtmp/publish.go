package rtmp

import (
	"fmt"
	"time"
	"context"
	"net/http"

	"github.com/zhangpeihao/goflv"
	rtmp "github.com/zhangpeihao/gortmp"
	"github.com/krislyy/qalive/configure"
)

const (
	programName = "RtmpPublisher"
	version     = "0.0.1"
)

type OutboundConnHandler struct {
	VideoDataSize 		int64
	AudioDataSize 		int64
	Status 				uint
	CreateStreamChan 	chan rtmp.OutboundStream
	CancelFunc 			context.CancelFunc
	ObConn 				rtmp.OutboundConn
	Config 				configure.Configure
}

func (handler *OutboundConnHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	if handler.ObConn == nil {
		return
	}
	handler.Status, err = handler.ObConn.Status()
	fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", handler.Status, err)
}

func (handler *OutboundConnHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")
	handler.CancelFunc()
}

func (handler *OutboundConnHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
}

func (handler *OutboundConnHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedRtmpCommand: %+v\n", command)
}

func (handler *OutboundConnHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	handler.CreateStreamChan <- stream
}
func (handler *OutboundConnHandler) OnPlayStart(stream rtmp.OutboundStream) {

}
func (handler *OutboundConnHandler) OnPublishStart(stream rtmp.OutboundStream) {
	// Set chunk buffer size
	go handler.publish(stream)
}

func (handler *OutboundConnHandler) publish(stream rtmp.OutboundStream) {
	var err error
	cacheTags := NewCacheTags(8, handler.Config)
	go cacheTags.StartCache()
	fmt.Println("2")
	startTs := uint32(0)
	startAt := time.Now().UnixNano()
	fmt.Println("3")
	for tag := range cacheTags.TagCh {
		if handler.Status != rtmp.OUTBOUND_CONN_STATUS_CREATE_STREAM_OK {
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
			handler.VideoDataSize += int64(len(tag.Data))
		case flv.AUDIO_TAG:
			handler.AudioDataSize += int64(len(tag.Data))
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
	//send httpGet stopTask api
	resp, err := http.Get("http://localhost:8081/api/stopTask?stream=" + handler.Config.StreamName)
	if err != nil {
		fmt.Println("http get error " + err.Error())
		return
	}
	defer resp.Body.Close()
}

type RTMP_Session struct {
	outHandler *OutboundConnHandler
	Finished   bool
}

func (rs *RTMP_Session) Publish(conf *configure.Configure)  {
	defer checkError()
	rs.outHandler = &OutboundConnHandler{
		CreateStreamChan: make(chan rtmp.OutboundStream),
		Config: *conf,
	}
	var err error
	rs.outHandler.ObConn, err = rtmp.Dial(rs.outHandler.Config.Crtmp_url, rs.outHandler, 100)
	if err != nil {
		panic(fmt.Sprintf("Dial error %s", err.Error()))
	}
	fmt.Println("b")
	defer rs.outHandler.ObConn.Close()
	fmt.Println("to connect")
	err = rs.outHandler.ObConn.Connect()
	if err != nil {
		panic(fmt.Sprintf("Connect error: %s", err.Error()))
	}
	fmt.Println("c")
	var ctx context.Context
	ctx, rs.outHandler.CancelFunc = context.WithCancel(context.Background())
	for {
		select {
		case stream := <-rs.outHandler.CreateStreamChan:
			// Publish
			stream.Attach(rs.outHandler)
			if rs.outHandler.Config.Params != "" {
				err = stream.Publish(rs.outHandler.Config.StreamName + "?" + rs.outHandler.Config.Params, "live")
			} else {
				err = stream.Publish(rs.outHandler.Config.StreamName, "live")
			}

			if err != nil {
				str := fmt.Sprintf("Publish error: %s", err.Error())
				panic(str)
			}

		case <-time.After(1 * time.Second):
			fmt.Printf("Audio size: %d bytes; Vedio size: %d bytes\n", rs.outHandler.AudioDataSize, rs.outHandler.VideoDataSize)

		case <-ctx.Done():
			fmt.Println("RTMP stream closed!")
			rs.Finished = true
			return
		}
	}
}

func (rs *RTMP_Session) Stop() {
	if !rs.Finished {
		rs.outHandler.CancelFunc()
	}
}

func checkError() {
	if e := recover(); e != nil {
		fmt.Println("RTMP_Session panic: ", e)
	}
}
