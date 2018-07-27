package core

import (
    "github.com/nareix/joy4/format"
    "github.com/nareix/joy4/format/rtmp"
	"github.com/krislyy/qalive/configure"
	"github.com/nareix/joy4/av"
)

func init() {
    rtmp.Debug = true
    format.RegisterAll()
}

func AsyncCopyPackets(configure *configure.Configure) {
	// srclist
	srclist := configure.GetSrcList()
	for _, v := range srclist {
		file := &VFile{Name: v}
		ch := make(chan av.Packet)
		headCh := make(chan []av.CodecData)
		// reader
		reader := CreateReader(ch, headCh, file)
		// writer
		writer := CreateWriter(ch, headCh, file)
		// start gorutines
		go reader.StartLoop()
		go writer.StartLoop()
		// wait
		<-reader.CloseCh
		writer.CloseCh <- true
	}
}

// check error and recover
func checkError() {
	defer func() {
		if e := recover(); e != nil {
		}
	}()
}