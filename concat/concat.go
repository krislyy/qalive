package concat

import (
	"fmt"
	"io"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/av/pktque"
	"log"
)

func init() {
	format.RegisterAll()
}

func ConcatMP4ToFLV(urls []string, temp_flv string) (err error){
	flv_muxer, err := avutil.Create(temp_flv)
	//TODO krilsyy fix time
	filter := &pktque.FixTime{StartFromZero: true, MakeIncrement: true}

	var writeHeader bool = false
	for i := 0; i< len(urls) ; i++ {
		err = func()(err error){ // for use defer
			file, err := avutil.Open(urls[i])
			if err != nil {
				log.Printf("Opne file error %s", err.Error())
				return
			}
			defer file.Close()

			fmt.Printf("File %d name: %s\n", i, urls[i])
			var streams []av.CodecData
			streams, err = file.Streams()
			if writeHeader == false {
				writeHeader = true
				if err = flv_muxer.WriteHeader(streams); err != nil {
					return
				}
			}

			for {
				var pkt av.Packet
				if pkt, err = file.ReadPacket(); err != nil {
					if err == io.EOF {
						break
					}
					return
				}
				filter.ModifyPacket(&pkt, nil, 0, 0)
				if err = flv_muxer.WritePacket(pkt); err != nil {
					return
				}
			}
			return
		}()
		if err != io.EOF {
			break
		}
	}
	if err = flv_muxer.WriteTrailer(); err != nil {
		return
	}
	flv_muxer.Close()
	return
}