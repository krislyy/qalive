package rtmp

import (
	"gome.com/qalive/configure"
	"github.com/zhangpeihao/goflv"
	"fmt"
)

type Tag struct {
	TagType   	byte
	Timestamp 	uint32
	Data 	  	[]byte
	IsFinished 	bool
}

type CacheTags struct {
	TagCh 	chan *Tag
	Idx 	int
	Files 	[]string
	Quit    bool
}

func NewCacheTags(n int, config configure.Configure) *CacheTags {
	return &CacheTags{
		TagCh:make(chan *Tag, n),
		Idx:-1,
		Files:config.GetPlayList(),
	}
}

func (c *CacheTags) getNextFile() string {
	nextIndex := c.Idx + 1
	if nextIndex >= len(c.Files) {
		nextIndex = 0
		return ""
	}
	c.Idx = nextIndex
	return c.Files[c.Idx]
}


func (c *CacheTags)StartCache() {
	for {
		if c.Quit {
			break
		}
		flvFile, err := flv.OpenFile(c.getNextFile())
		if err != nil {
			fmt.Println("Open FLV dump file error:", err)
			break
		}
		for {
			if c.Quit {
				break
			}
			tag := &Tag{ IsFinished:false }
			if flvFile.IsFinished() {
				tag.IsFinished = true
				c.TagCh <- tag
				break
			}
			header, data, err := flvFile.ReadTag()
			if err != nil {
				fmt.Println("flvFile.ReadTag() error:", err)
				break
			}
			tag.Timestamp = header.Timestamp
			tag.TagType = header.TagType
			tag.Data = data
			c.TagCh <- tag
		}
		flvFile.Close()
	}
	close(c.TagCh)
	return
}