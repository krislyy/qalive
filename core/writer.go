package core

import (
    "fmt"

    "github.com/nareix/joy4/av"
)

func CreateWriter(ch chan av.Packet, headCh chan []av.CodecData, file *VFile) *Writer {
    writer := &Writer{
        Ch:          ch,
        HeadCh:      headCh,
		CloseCh:      make(chan bool),
        File:        file,
    }
    return writer
}

// Writer is struct of writer to a muxer
type Writer struct {
    Ch          chan av.Packet
    HeadCh      chan []av.CodecData
	CloseCh      chan bool
    Destination av.MuxCloser
    File        *VFile
}

// StartLoop is function for starting listen chan of packets and write they to muxer
func (wr *Writer) StartLoop() {
    for {
        select {
        case hDat := <-wr.HeadCh:
        	wr.Destination, _ = wr.File.GetMuxer()
            fmt.Println(wr.File.Name, " WriteHeader")
            wr.Destination.WriteHeader(hDat)

        case pkg := <-wr.Ch:
            err := wr.Destination.WritePacket(pkg)
            if err != nil {
                fmt.Println("Error on write packet to Destination", err)
            }

		case <-wr.CloseCh:
			fmt.Println(wr.File.Name, " WriteTailer")
			wr.Destination.WriteTrailer()
			wr.Destination.Close()
        	return

        default:

        }
    }
}