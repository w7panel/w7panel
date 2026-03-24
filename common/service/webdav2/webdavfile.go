package webdav2

import (
	"encoding/xml"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/net/webdav"
)

type WebDAVFile struct {
	webdav.File
	*WebDAVFileInfo
}

type WebDAVFileInfo struct {
	os.FileInfo
}

func (n *WebDAVFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	stat, err := n.Stat()
	if err != nil {
		return nil, err
	}
	user := webdav.Property{
		XMLName:  xml.Name{Local: "uid", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	group := webdav.Property{
		XMLName:  xml.Name{Local: "gid", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	perm := webdav.Property{
		XMLName:  xml.Name{Local: "mode", Space: "w7panel"},
		InnerXML: []byte("unknown"),
	}
	filestat, ok := stat.Sys().(*syscall.Stat_t)
	if ok {
		user.InnerXML = []byte(fmt.Sprintf("%d", filestat.Uid))
		group.InnerXML = []byte(fmt.Sprintf("%d", filestat.Gid))
		perm.InnerXML = []byte(fmt.Sprintf("%o", stat.Mode().Perm()))
	}
	ret := make(map[xml.Name]webdav.Property)
	ret[user.XMLName] = user
	ret[group.XMLName] = group
	ret[perm.XMLName] = perm
	return ret, nil
}

func (n *WebDAVFile) Stat() (os.FileInfo, error) {
	if n.WebDAVFileInfo != nil {
		return n.WebDAVFileInfo, nil
	}
	stat, err := n.File.Stat()
	if err != nil {
		return nil, err
	}
	n.WebDAVFileInfo = &WebDAVFileInfo{stat}
	return n.WebDAVFileInfo, nil
}
