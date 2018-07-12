package core

import (
    "fmt"
    "github.com/nareix/joy4/av"
)

// CreateReader return new Reader with chanel
func CreateReader(ch chan av.Packet, headCh chan []av.CodecData, file *VFile) *Reader {
    r := &Reader{
        Ch:      ch,
        HeadCh:  headCh,
        CloseCh: make(chan bool),
        File:   file,
    }
    return r
}

// Reader is
type Reader struct {
    Ch      chan av.Packet
    HeadCh  chan []av.CodecData
    CloseCh chan bool
    File    *VFile
}

func (r *Reader) StartLoop() {
	demuxer, err := r.File.GetDemuxer()
	if err != nil {
		fmt.Printf("Error on getting demuxer from VFile %s \n", r.File)
		fmt.Println(err)
		r.CloseCh <- true
		return
	}
	defer demuxer.Close()
	// Write headers
	codecDat, err := demuxer.Streams()
	if err != nil {
		fmt.Printf("Error on getting streams from Demuxer %s \n", r.File)
		fmt.Println(err)
		r.CloseCh <- true
		return
	}
	r.HeadCh <- codecDat
	// Write packets
	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			fmt.Println("Error on getting packet;", err)
			break
		}
		r.Ch <- pkg
	}
	r.CloseCh <- true
}
