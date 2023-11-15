package gopub

type Container struct {
	RootFile RootFile `xml:"rootfiles>rootfile"`
}

type RootFile struct {
	FullPath string `xml:"full-path,attr"`
}
