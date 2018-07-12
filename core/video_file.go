package core

import (
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"fmt"
)

type VFile struct {
	Name string
}

// GetDemuxer is function for getting demuxer from file
func (vf *VFile) GetDemuxer() (av.DemuxCloser, error) {
	file, err := avutil.Open(vf.Name + ".mp4")
	if err != nil {
		fmt.Println("Error on open mp4 file;", err.Error())
		return nil, err
	}
	return file, err
}

func (vf *VFile) GetMuxer() (av.MuxCloser, error) {
	file, err := avutil.Create(vf.Name + ".flv")
	if err != nil {
		fmt.Println("Error on create flv file;", err.Error())
		return nil, err
	}
	return file, err
}
