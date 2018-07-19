package configure

func NewConfigure() *Configure{
	return &Configure{
		DefaultDir: "/Users/kaolafm/Desktop/",
	}
}

type Configure struct {
	DefaultDir 		string `json:"default_dir"`
	SrcList		[]string `json:"srclist`
	PlayList  	[]string `json:"playlist`
	Crtmp_url 	string   `json:"rtmp_url"`
	StreamName 	string   `json:"streamName"`
	Token		string   `json:"token"`
}

func (self *Configure)GetPlayList() []string {
	for key, value := range self.PlayList {
		self.PlayList[key] = self.DefaultDir + self.StreamName + "/" +value + ".flv"
	}
	return self.PlayList[:]
}

func (self *Configure)GetSrcList() []string {
	for key, value := range self.SrcList {
		self.SrcList[key] = self.DefaultDir + self.StreamName + "/" +value
	}
	return self.SrcList[:]
}

// obsolete
func (self *Configure)GetTempFlvPath() string {
	return self.DefaultDir + self.StreamName + "/temp.flv"
}

func (self *Configure)IsPushValid() bool {
	return len(self.PlayList) > 0 && self.Crtmp_url != "" &&
		self.DefaultDir != "" && self.StreamName != ""
}

func (self *Configure)IsCopyValid() bool {
	return len(self.SrcList) > 0 && self.DefaultDir != "" && self.StreamName != ""
}
