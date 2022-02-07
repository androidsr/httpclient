package qcode

import (
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/walk"
	goqrcode "github.com/skip2/go-qrcode"
	"os"
	"fmt"
	"github.com/tuotoo/qrcode"
)

type QrCode struct {
	form   *walk.Dialog
	qrImg  *walk.ImageView
	qrData *walk.LineEdit
}

func (q *QrCode) Show() {

	Dialog{
		AssignTo: &q.form,
		Title:    "二维码",
		MinSize:  Size{400, 500},
		MaxSize:  Size{400, 500},
		Layout:   Grid{Columns: 1,},
		Children: []Widget{
			Label{
				Text:    "二维码数据",
				MaxSize: Size{0, 25},
			},
			LineEdit{
				AssignTo: &q.qrData,
			},
			Composite{
				Layout: Grid{Columns: 2,},
				Children: []Widget{
					PushButton{
						Text: "识别二维码",
						OnClicked: func() {
							d := new(walk.FileDialog)
							d.Title = "选择二维码"
							if f, _ := d.ShowOpen(q.form); !f {
								return
							}
							fi, err := os.Open(d.FilePath)
							if err != nil {
								fmt.Println(err.Error())
								return
							}
							defer fi.Close()
							qrmatrix, err := qrcode.Decode(fi)
							if err != nil {
								fmt.Println(err.Error())
								return
							}
							q.qrData.SetText(qrmatrix.Content)

						},
					},
					PushButton{
						Text: "生成二维码",
						OnClicked: func() {
							if q.qrData.Text() == "" {
								return
							}
							goqrcode.WriteFile(q.qrData.Text(), goqrcode.Medium, 350, "qrcode.png")
							img, _ := walk.NewImageFromFile("qrcode.png")
							q.qrImg.SetImage(img)
						},
					},
				},
			},
			ImageView{
				AssignTo: &q.qrImg,
			},
		},
	}.Run(nil)
}
