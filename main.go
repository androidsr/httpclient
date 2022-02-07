package main


import (
	"github.com/braintree/manners"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"tools/http"
	"tools/qcode"
	"tools/tcp"
)

func main() {
	var mw *walk.MainWindow
	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "网络工具",
		Size:     Size{250, 120},
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				Text: "HTTP客户端",
				OnClicked: func() {
					c := http.HttpCli{}
					c.Show()
				},
			},
			VSpacer{},
			PushButton{
				Text: "HTTP服务",
				OnClicked: func() {
					h := http.HttpServ{}
					h.Show()
					if h.ServFlag {
						manners.Close()
					}
				},
			},
			VSpacer{},
			PushButton{
				Text: "TCP客户端",
				OnClicked: func() {
					t := tcp.TcpCli{}
					t.Show()
				},
			},
			VSpacer{},
			PushButton{
				Text: "TCP服务",
				OnClicked: func() {
					t := tcp.TcpServ{}
					t.Show()
					if t.ServFlag {
						t.Close()
					}

				},
			},
			VSpacer{},
			PushButton{
				Text: "二维码工具",
				OnClicked: func() {
					q := qcode.QrCode{}
					q.Show()
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
