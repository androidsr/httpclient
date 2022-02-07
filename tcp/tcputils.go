package tcp

import (
	"encoding/json"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"net"
	"tools/http"
)

var (
	DB = http.DB
)

type TcpCli struct {
	form  *walk.MainWindow
	addr  *walk.LineEdit
	read  *walk.TextEdit
	write *walk.TextEdit
	btn   *walk.PushButton
	data  *TcpCliData
}

func (t *TcpCli) Show() {
	bs, _ := DB.Get([]byte("tcp.client.data"), nil)
	t.data = new(TcpCliData)
	json.Unmarshal(bs, t.data)

	MainWindow{
		AssignTo: &t.form,
		Title:    "TCP服务器",
		MinSize:  Size{600, 500},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{
						Text: "连接地址",
					},
					LineEdit{
						AssignTo: &t.addr,
						Text:     t.data.Addr,
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2, MarginsZero: true},
				Children: []Widget{
					Label{
						Text:       "发送",
						ColumnSpan: 2,
					},
					TextEdit{
						AssignTo: &t.write,
					},

					Label{
						Text:       "响应",
						ColumnSpan: 2,
					},
					TextEdit{
						AssignTo: &t.read,
						Text:     t.data.Read,
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						AssignTo: &t.btn,
						Text:     "发送",
						OnClicked: func() {
							go t.connect()
						},
					},
				},
			},
		},
	}.Run()
}

type TcpCliData struct {
	Addr  string
	Read  string
	Write string
}

func (t *TcpCli) connect() {
	t.data.Addr = t.addr.Text()
	t.data.Write = t.write.Text()
	if t.data.Addr == "" {
		return
	}

	addr, _ := net.ResolveTCPAddr("tcp", t.data.Addr)
	var err error
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		walk.MsgBox(t.form, "错误", "连接失败", walk.MsgBoxOK)
		return
	}
	defer conn.Close()
	conn.Write([]byte(t.write.Text()))
	conn.CloseWrite()
	bs, err := ioutil.ReadAll(conn)
	if err != nil {
		return
	}
	t.read.SetText(t.read.Text() + "\r\n" + string(bs))
	wd, _ := json.Marshal(t.data)
	DB.Put([]byte("tcp.client.data"), wd, nil)
}

//================================TCP服务器==============================
type TcpServ struct {
	ServFlag bool
	listen   *net.TCPListener
	form     *walk.MainWindow
	port     *walk.LineEdit
	read     *walk.TextEdit
	write    *walk.TextEdit
	btn      *walk.PushButton
	data     *TcpServData
}

func (t *TcpServ) Show() {
	bs, _ := DB.Get([]byte("tcp.server.data"), nil)
	t.data = new(TcpServData)
	json.Unmarshal(bs, t.data)
	MainWindow{
		AssignTo: &t.form,
		Title:    "TCP服务器",
		MinSize:  Size{600, 500},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{
						Text: "端口",
					},
					LineEdit{
						AssignTo: &t.port,
						Text:     t.data.Port,
					},
				},
			},
			Composite{
				Layout: Grid{Columns: 2, MarginsZero: true},
				Children: []Widget{
					Label{
						Text:       "接收",
						ColumnSpan: 2,
					},
					TextEdit{
						AssignTo: &t.read,
						Text:     t.data.Read,
					},
					Label{
						Text:       "返回",
						ColumnSpan: 2,
					},
					TextEdit{
						AssignTo: &t.write,
						Text:     t.data.Write,
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						AssignTo: &t.btn,
						Text:     "启动",
						OnClicked: func() {
							t.runServer()
						},
					},
				},
			},
		},
	}.Run()
}

type TcpServData struct {
	Port  string
	Read  string
	Write string
}

func (t *TcpServ) runServer() {
	t.data.Port = t.port.Text()
	t.data.Read = t.read.Text()
	t.data.Write = t.write.Text()
	if !t.ServFlag {
		if t.data.Port == "" {
			return
		}
		addr, _ := net.ResolveTCPAddr("tcp", ":"+t.data.Port)
		var err error
		t.listen, err = net.ListenTCP("tcp", addr)
		if err != nil {
			walk.MsgBox(t.form, "错误", "启动服务失败", walk.MsgBoxOK)
			return
		}
		go func() {
			for {
				conn, err := t.listen.AcceptTCP()
				if err != nil {
					continue
				}
				go func() {
					defer conn.Close()
					bs, err := ioutil.ReadAll(conn)
					if err != nil {
						return
					}
					t.read.SetText(t.read.Text() + "\r\n" + string(bs))
					conn.Write([]byte(t.write.Text()))
				}()
			}
		}()
		t.btn.SetText("停止")
		wd, _ := json.Marshal(t.data)
		DB.Put([]byte("tcp.server.data"), wd, nil)
	} else {
		t.btn.SetText("启动")
		t.Close()
	}
	t.ServFlag = !t.ServFlag
}

func (t *TcpServ) Close() {
	t.listen.Close()
}
