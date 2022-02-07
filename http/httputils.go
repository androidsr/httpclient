package http

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/braintree/manners"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/skip2/go-qrcode"
	"github.com/srlemon/gen-id/generator"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	DB         *leveldb.DB
	cTypeModel []string
	mTypeModel []string
	httpData   = bytes.Buffer{}
	headerMap  = make(map[string]string, 0)
	cookieMap  = make(map[string]string, 0)
)

//	go build -ldflags="-H windowsgui"
const (
	RN         = "\r\n"
	RAND_STR   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	INPUT_DATA = "url = http://localhost:8080" + RN +
		" #提交方式【 POST GET 】" + RN +
		"method = POST" + RN +
		" #数据格式【application/x-www-form-urlencoded / application/json / text/xml / text/html / multipart/form-data】" + RN +
		"headers.Content-Type =  application/json" + RN +
		" #暂停时间（秒）" + RN +
		"sleep = 0 " + RN +
		" #随机数字生成（流水号）" + RN +
		"number = int:999999" + RN +
		" #随机字符串" + RN +
		"strvalue = str:32" + RN +
		" #19位长度日期时间" + RN +
		"date1 = date:19" + RN +
		" #14位长度日期时间" + RN +
		"date1 = date:14" + RN +
		" #8位长度日期" + RN +
		"date2 = date:8" + RN +
		" #6位长度时间" + RN +
		"date3 = date:6" + RN +
		" #从指定位置请求报文中的指定位置取值" + RN +
		"up_value = req:0:txn_no" + RN +
		" #从响应报文中查找值，格式：res:(开头):0(从第0个交易结果中取值):<a>||</a>(以||分划取查找开始位置||结束位置)" + RN +
		"result_val = res:0:<a>||</a>" + RN +
		" #报文结束符号" + RN +
		"=END="
)

func init() {
	DB, _ = leveldb.OpenFile("./db", nil)
	cTypeModel = []string{"application/json", "application/x-www-form-urlencoded", "text/plain", "text/xml", "text/html"}
	sort.Strings(cTypeModel)
	mTypeModel = []string{"POST", "GET"}
	sort.Strings(mTypeModel)
}

type HttpCli struct {
	form      *walk.MainWindow
	name      *walk.ComboBox
	input     *walk.TextEdit
	resultAll *walk.TextEdit
	fTime     *walk.NumberEdit
	start     *walk.NumberEdit
	end       *walk.NumberEdit
	forCheck  *walk.CheckBox
	sendType  *walk.CheckBox
	click     *walk.PushButton
	qrImg     *walk.ImageView
	qrKey     *walk.LineEdit
	fThread   *walk.NumberEdit
}

type CliData struct {
	inputGroup []map[string]interface{}
	result     []string
	Name       string          `json:"name"`
	Input      string          `json:"input"`
	ForCheck   walk.CheckState `json:"for_check"`
	SendType   walk.CheckState `json:"send_type"`
	ForTime    float64         `json:"for_time"`
	Start      float64         `json:"start"`
	QrKey      string          `json:"qr_key"`
	End        float64         `json:"end"`
	FThread    float64         `json:"f_thread"`
}

func (d *CliData) WriteDB() {
	bs, _ := json.Marshal(d)
	fmt.Println(string(bs))
	DB.Put([]byte(d.Name), bs, nil)
}

func (d *CliData) ReadAllDB() []CliData {
	iter := DB.NewIterator(nil, nil)
	var datas []CliData
	for iter.Next() {
		bs := iter.Value()
		data := CliData{}
		json.Unmarshal(bs, &data)
		if data.Name == "" {
			continue
		}
		datas = append(datas, data)
	}
	return datas
}

func (d *CliData) DelDB(key string) {
	DB.Delete([]byte(key), nil)
}

func (h *HttpCli) readData() *CliData {
	buf := bufio.NewReader(strings.NewReader(h.input.Text()))
	data := new(CliData)
	var item = make(map[string]interface{})
	data.inputGroup = []map[string]interface{}{}
	headers := make(map[string]string, 0)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "\r\n" {
			continue
		}
		if strings.HasPrefix(line, "headers.") {
			idx := strings.Index(line, ".") + 1
			le := line[idx:]
			headers[strings.TrimSpace(le[0:strings.Index(le, "=")])] = strings.TrimRight(strings.TrimLeft(le[strings.Index(le, "=")+1:], " "), " ")
			continue
		}
		//报文分组标志
		if strings.Contains(line, "=END=") {
			item["headers"] = headers
			data.inputGroup = append(data.inputGroup, item)
			item = make(map[string]interface{})
			headers = make(map[string]string, 0)
			continue
		}

		if !strings.Contains(line, "=") {
			if h.sendType.CheckState() == 1 {
				val := item["data"]
				if val != nil {
					item["data"] = val.(string) + line
				}
			}
			continue
		}

		lines := [2]string{line[0:strings.Index(line, "=")], line[strings.Index(line, "=")+1:]}
		val := strings.TrimSpace(lines[1])
		comp := regexp.MustCompile(`idCart:[\d]+`)
		cards := comp.FindAllString(val, -1)
		idNames := make(map[string]string, len(cards))
		for _, oldStr := range cards {
			comp = regexp.MustCompile(`[\d]+`)
			num := comp.FindString(oldStr)
			g := new(generator.GeneratorData)
			idNames[fmt.Sprint("idName:", num)] = g.GeneratorName()
			idCard, _, _, _, _ := g.GeneratorIDCart()
			newValue := strings.ReplaceAll(val, oldStr, idCard)
			val = newValue
		}

		comp = regexp.MustCompile(`idName:[\d]+`)
		names := comp.FindAllString(val, -1)
		for _, oldStr := range names {
			comp = regexp.MustCompile(`[\d]+`)
			num := comp.FindString(oldStr)
			name := idNames[fmt.Sprint("idName:", num)]
			if name == "" {
				g := new(generator.GeneratorData)
				name = g.GeneratorName()
			}
			newValue := strings.ReplaceAll(val, oldStr, name)
			val = newValue
		}

		comp = regexp.MustCompile(`int:[\d]+`)
		idxs := comp.FindIndex([]byte(val))
		if len(idxs) == 2 {
			f := val[idxs[0]:idxs[1]]
			n, _ := strconv.Atoi(f[4:])
			rand.Seed(time.Now().UnixNano())
			nVal := fmt.Sprint(rand.Intn(n))
			val = strings.ReplaceAll(val, f, nVal)
		}

		comp = regexp.MustCompile(`date:[\d]+`)
		idxs = comp.FindIndex([]byte(val))
		if len(idxs) == 2 {
			f := val[idxs[0]:idxs[1]]
			f = f[5:]
			var nVal string
			if f == "8" {
				nVal = time.Now().Format("20060102")
			} else if f == "14" {
				nVal = time.Now().Format("20060102150405")
			} else if f == "19" {
				nVal = time.Now().Format("2006-01-02 15:04:05")
			} else if f == "6" {
				nVal = time.Now().Format("150405")
			} else if f == "10" {
				nVal = time.Now().Format("2006-01-02")
			} else if f == "timestamp" {
				nVal = fmt.Sprint(time.Now().Unix())
			}
			o := "date:" + f
			val = strings.ReplaceAll(val, o, nVal)
		}

		comp = regexp.MustCompile(`str:[\d]+`)
		idxs = comp.FindIndex([]byte(val))
		if len(idxs) == 2 {
			f := val[idxs[0]:idxs[1]]
			length, _ := strconv.Atoi(f[4:])
			nVal := randomStr(length)
			val = strings.ReplaceAll(val, f, nVal)
		}
		item[strings.TrimSpace(lines[0])] = val
	}
	if len(item) != 0 {
		item["headers"] = headers
		data.inputGroup = append(data.inputGroup, item)
	}
	data.Start = h.start.Value()
	data.End = h.end.Value()
	data.Name = h.name.Text()
	data.Input = h.input.Text()
	data.ForCheck = h.forCheck.CheckState()
	data.SendType = h.sendType.CheckState()
	data.ForTime = h.fTime.Value()
	data.FThread = h.fThread.Value()
	data.result = []string{}
	data.QrKey = h.qrKey.Text()

	if (int(data.Start) < len(data.inputGroup) && int(data.Start) != 0) || int(data.End) < len(data.inputGroup) {
		var task []map[string]interface{}
		max := math.Min(data.End, float64(len(data.inputGroup)))
		for i := int(data.Start); i < int(max); i++ {
			task = append(task, data.inputGroup[i])
		}
		data.inputGroup = task
	}
	return data
}

func (h *HttpCli) bindData(data CliData) {
	h.name.SetText(data.Name)
	h.forCheck.SetCheckState(data.ForCheck)
	h.sendType.SetCheckState(data.SendType)
	h.fTime.SetValue(data.ForTime)
	h.fThread.SetValue(data.FThread)
	h.input.SetText(data.Input)
	h.qrKey.SetText(data.QrKey)
}

func (h *HttpCli) Click() {
	httpData = bytes.Buffer{}

	if h.forCheck.Checked() {
		ch := make(chan int, int(h.fThread.Value()))
		go func() {
			defer func() {
				recover()
			}()
			httpData.WriteString("处理中...")
			h.Println()
			for {
				ch <- 1
				go h.sendHttp(ch)
				if !h.forCheck.Checked() {
					name := "D:/" + time.Now().Format("20060102150405") + ".log"
					os.WriteFile(name, httpData.Bytes(), 0777)
					h.resultAll.SetText("请查看文件：" + name)
					break
				}
				a := time.Duration(h.fTime.Value()) * time.Millisecond
				time.Sleep(a)
			}
		}()
	} else {
		httpData.WriteString("处理中...")
		h.Println()
		ch := make(chan int, 1)
		ch <- 1
		go h.sendHttp(ch)
		go func() {
			ch <- 1
			h.Println()
		}()
	}
}

func (h *HttpCli) sendHttp(ch chan int) {
	defer func() {
		<-ch
	}()
	data := h.readData()
	for _, item := range data.inputGroup {
		h.itemChange(data, item)
		client := &http.Client{}
		var resp *http.Response
		sleep, ok := item["sleep"].(string)

		if ok {
			delete(item, "sleep")
			out, _ := strconv.Atoi(sleep)
			time.Sleep(time.Second * time.Duration(out))
		}
		headers := item["headers"].(map[string]string)
		delete(item, "headers")

		addr, ok := item["url"].(string)
		if !ok {
			walk.MsgBox(h.form, "警告", "URL不能为空！", walk.MsgBoxOK)
			return
		}
		delete(item, "url")
		method, ok := item["method"].(string)
		if !ok {
			walk.MsgBox(h.form, "警告", "请求类型不能为空！", walk.MsgBoxOK)
			return
		}
		delete(item, "method")
		if addr == "" {
			walk.MsgBox(h.form, "警告", "URL不能为空！", walk.MsgBoxOK)
			return
		}
		encryptKey, _ := item["encryptKey"].(string)
		salt, _ := item["salt"].(string)
		delete(item, "encryptKey")
		delete(item, "salt")

		h.Append("请求地址：" + addr)
		sendData, ok := item["data"].(string)
		if h.sendType.CheckState() == 1 && ok && sendData != "" {
			hs, _ := json.Marshal(headers)
			h.Append("Headers：" + string(hs))
			h.Append(sendData)

			req, err := http.NewRequest(method, addr, strings.NewReader(sendData))
			for k, v := range headerMap {
				req.Header.Add(k, v)
			}
			for k, v := range cookieMap {
				ck := &http.Cookie{Name: k, Value: v, HttpOnly: true}
				req.AddCookie(ck)
			}
			for k, v := range headers {
				req.Header.Add(k, v)
			}

			resp, err = client.Do(req)
			if err != nil {
				h.PrintError(err)
				return
			}
			defer resp.Body.Close()
		} else {
			if strings.Contains(headers["Content-Type"], "application/json") {
				nItem := h.sign(cloneMaps(item), encryptKey, salt)
				bs, err := json.Marshal(nItem)
				if err != nil {
					h.PrintError(err)
					return
				}
				hs, _ := json.Marshal(headers)
				h.Append("Headers：" + string(hs))
				h.Append("请求数据：" + string(bs))

				req, err := http.NewRequest(method, addr, bytes.NewBuffer(bs))
				for k, v := range headerMap {
					req.Header.Add(k, v)
				}
				for k, v := range cookieMap {
					ck := &http.Cookie{Name: k, Value: v, HttpOnly: true}
					req.AddCookie(ck)
				}
				for k, v := range headers {
					req.Header.Add(k, v)
				}
				resp, err = client.Do(req)

				if err != nil {
					h.PrintError(err)
					return
				}
				defer resp.Body.Close()
			} else if strings.Contains(headers["Content-Type"], "multipart/form-data") {
				uploadName, ok := item["uploadName"].(string)
				if !ok {
					h.PrintError(errors.New("上传文件时：[uploadName]标签名不能为空"))
					return
				}
				delete(item, "uploadName")
				filePath, ok := item["filePath"].(string)
				if !ok {
					h.PrintError(errors.New("上传文件时：[filePath]文件路径不能为空"))
					return
				}
				delete(item, "filePath")

				buf := new(bytes.Buffer)
				bodyWriter := multipart.NewWriter(buf) // body writer

				f, err := os.Open(filePath)
				if err != nil {
					return
				}
				defer f.Close()

				for k, v := range item {
					vv, ok := v.(string)
					if ok {
						bodyWriter.WriteField(k, vv)
					}
				}
				fw, _ := bodyWriter.CreateFormFile(uploadName, f.Name())
				io.Copy(fw, f)
				bodyWriter.Close()
				contentType := bodyWriter.FormDataContentType()
				req, err := http.NewRequest(method, addr, buf)

				// bs, err := json.Marshal(item)
				// if err != nil {
				// 	h.PrintError(err)
				// 	return
				// }
				hs, _ := json.Marshal(headers)
				h.Append("Headers：" + string(hs))
				for k, v := range headerMap {
					req.Header.Add(k, v)
				}
				for k, v := range cookieMap {
					ck := &http.Cookie{Name: k, Value: v, HttpOnly: true}
					req.AddCookie(ck)
				}
				for k, v := range headers {
					req.Header.Add(k, v)
				}
				req.Header.Set("Content-Type", contentType)

				for k, v := range headers {
					req.Header.Add(k, v)
				}
				resp, err = client.Do(req)
				if err != nil {
					h.PrintError(err)
					return
				}
				defer resp.Body.Close()

			} else {
				sdata := url.Values{}
				for k, v := range item {
					val, _ := v.(string)
					sdata.Add(k, val)
				}
				hs, _ := json.Marshal(headers)
				h.Append("Headers：" + string(hs))
				h.Append("请求数据：" + sdata.Encode())
				var req *http.Request
				var err error
				if strings.ToUpper(method) == "GET" {
					if strings.Contains(addr, "?") {
						addr = addr + "&" + sdata.Encode()
					} else {
						addr = addr + "?" + sdata.Encode()
					}
					req, err = http.NewRequest(method, addr, nil)
				} else {
					req, err = http.NewRequest(method, addr, strings.NewReader(sdata.Encode()))
				}

				for k, v := range headerMap {
					req.Header.Add(k, v)
				}
				for k, v := range cookieMap {
					ck := &http.Cookie{Name: k, Value: v, HttpOnly: true}
					req.AddCookie(ck)
				}
				for k, v := range headers {
					req.Header.Add(k, v)
				}
				resp, err = client.Do(req)
				if err != nil {
					h.PrintError(err)
					return
				}
				defer resp.Body.Close()
			}
		}
		for k, v := range resp.Header {
			if k == "Date" || k == "Vary" || k == "Content-Type" {
				continue
			}
			if len(v) >= 1 {
				headerMap[k] = v[0]
			}
		}
		for _, v := range resp.Cookies() {
			cookieMap[v.Name] = v.Value
		}
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			h.PrintError(err)
			return
		}
		disp := resp.Header["Content-Disposition"]
		if len(disp) == 1 && strings.Contains(strings.ToLower(disp[0]), "filename") {
			u, _ := user.Current()
			start := strings.Index(disp[0], `"`)
			end := strings.LastIndex(disp[0], `"`)
			filename := disp[0]
			if start != -1 && end != -1 {
				filename = disp[0][start+1 : end]
			}

			ioutil.WriteFile(u.HomeDir+"/Downloads/"+filename, bs, 0666)
			h.Append("文件已保存：" + u.HomeDir + "/Downloads/" + filename)
		} else {
			h.Append("响应信息：")
			resultMap := make(map[string]interface{}, 0)
			err := json.Unmarshal(bs, &resultMap)
			var result string
			if err == nil && salt != "" && encryptKey != "" {
				unSign(resultMap, encryptKey, salt)
				retStr, err := json.Marshal(resultMap)
				if err != nil {
					h.Append(err.Error())
				}
				result = string(retStr)
			} else {
				result = string(bs)
			}

			h.Append(result)
			data.result = append(data.result, result)

			se := strings.Split(h.qrKey.Text(), "||")
			if len(se) == 2 {
				reg := regexp.MustCompile(se[0] + `.*` + se[1])
				finds := reg.FindAllString(result, -1)
				if len(finds) != 0 {
					find := finds[0]
					qrcode.WriteFile(find[strings.Index(find, se[0])+len(se[0]):strings.Index(find, se[1])], qrcode.Medium, 200, "qr.png")
					img, _ := walk.NewImageFromFile("qr.png")
					h.qrImg.SetImage(img)
				}
			}
		}

		h.Append(RN)
	}
}

func cloneMaps(tags map[string]interface{}) map[string]interface{} {
	cloneTags := make(map[string]interface{})
	for k, v := range tags {
		cloneTags[k] = v
	}
	return cloneTags
}

//打印错误
func (h *HttpCli) PrintError(err error) {
	walk.MsgBox(h.form, "错误", err.Error(), walk.MsgBoxOK)
}

//加签
func (h *HttpCli) sign(dataMap map[string]interface{}, encryptKey string, salt string) map[string]interface{} {
	requestData, ok := dataMap["requestData"].(string)
	if !ok {
		return dataMap
	}
	if requestData != "" && encryptKey != "" && salt != "" { //加签
		var buf bytes.Buffer
		aestool := AesCryptor{Key: []byte(encryptKey)}
		data, err := aestool.Encrypt([]byte(requestData))
		if err != nil {
			return dataMap
		}
		dataMap["requestData"] = base64.StdEncoding.EncodeToString(data)
		buf.WriteString(dataMap["requestData"].(string))
		buf.WriteString("|")
		buf.WriteString(salt)

		sysSign := fmt.Sprintf("%x", md5.Sum([]byte(buf.String())))
		dataMap["sysSign"] = sysSign
		h.Append("加密源数据：" + requestData)
		h.Append(RN)
	}
	return dataMap
}

//解密
func unSign(dataMap map[string]interface{}, encryptKey string, salt string) {
	responseData, ok := dataMap["responseData"].(string)
	if !ok {
		return
	}
	if responseData != "" && encryptKey != "" && salt != "" { //加签
		aestool := AesCryptor{Key: []byte(encryptKey)}
		bs, err := base64.StdEncoding.DecodeString(responseData)
		if err != nil {
			return
		}
		data, err := aestool.Decrypt(bs)
		if err != nil {
			return
		}
		dataMap["responseData"] = string(data)
	}
}

//RandomStr 随机生成字符串
func randomStr(size int) string {
	bytes := []byte(RAND_STR)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < size; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//打印提示消息
func (h *HttpCli) Append(text string) {
	httpData.WriteString(text)
	httpData.WriteString(RN)

}

func (h *HttpCli) Println() {
	h.resultAll.SetText(httpData.String())
	httpData = bytes.Buffer{}
}

//数据读取到map
func (h *HttpCli) itemChange(data *CliData, item map[string]interface{}) []string {
	var keys []string
	for k, v := range item {
		keys = append(keys, k)
		strVal, ok := v.(string)
		if !ok {
			continue
		}
		comp := regexp.MustCompile(`req:[0-9]+[:]+[a-zA-Z0-9\.]+`)
		aVals := comp.FindAllString(strVal, -1)
		for _, im := range aVals {
			vs := strings.Split(im, ":")
			if len(vs) != 3 {
				continue
			}
			var i, _ = strconv.Atoi(vs[1])
			val := vs[2]
			if strings.Contains(val, ".") {
				ks := strings.Split(val, ".")
				var inter = data.inputGroup[i]
				var retval string
				for l, s := range ks {
					if inter == nil {
						break
					}
					itv := inter[s]
					if itv == nil {
						inter = nil
						break
					}

					if (l + 1) != len(ks) {
						tiv, ok := itv.(string)
						if ok {
							var vMap map[string]interface{}
							err := json.Unmarshal([]byte(tiv), &vMap)
							if err == nil {
								inter = vMap
							}
						} else {
							miv, ok := itv.(map[string]interface{})
							if ok {
								inter = miv
							}
						}
						continue
					} else {
						retval, ok = itv.(string)
						if !ok {
							retbs, err := json.Marshal(itv)
							if err == nil {
								retval = string(retbs)
							}
						}
					}
				}
				strVal = strings.ReplaceAll(strVal, im, retval)
			} else {
				itmv := data.inputGroup[i][val]
				if itmv != nil {
					strVal = strings.ReplaceAll(strVal, im, itmv.(string))
				}
			}
		}

		comp = regexp.MustCompile(`res:[0-9]+[:]+[a-zA-Z0-9\.]+`)
		aVals = comp.FindAllString(strVal, -1)
		for _, im := range aVals {
			vs := strings.Split(im, ":")
			if len(vs) != 3 {
				continue
			}
			var i, _ = strconv.Atoi(vs[1])
			if len(data.result) <= i {
				continue
			}
			var inter map[string]interface{}
			err := json.Unmarshal([]byte(data.result[i]), &inter)
			if err != nil {
				continue
			}
			val := vs[2]
			if strings.Contains(val, ".") {
				ks := strings.Split(val, ".")
				var retval string
				for i, s := range ks {
					if inter == nil {
						break
					}
					itv := inter[s]
					if itv == nil {
						inter = nil
						break
					}

					if (i + 1) != len(ks) {
						var vMap map[string]interface{}
						err := json.Unmarshal([]byte(itv.(string)), &vMap)
						if err == nil {
							inter = vMap
						}
						continue
					} else {
						retval, ok = itv.(string)
						if !ok {
							retbs, err := json.Marshal(itv)
							if err == nil {
								retval = string(retbs)
							}
						}
					}
				}
				strVal = strings.ReplaceAll(strVal, im, retval)
			} else {
				itmv := data.inputGroup[i][val]
				if itmv != nil {
					strVal = strings.ReplaceAll(strVal, im, itmv.(string))
				}
			}
		}
		item[k] = strVal
	}
	sort.Strings(keys)
	return keys
}

func (h *HttpCli) Show() {
	data := new(CliData)

	MainWindow{
		AssignTo: &h.form,
		Title:    "HTTP客户端",
		MinSize:  Size{1000, 700},
		Layout:   HBox{},
		Children: []Widget{
			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					TextEdit{
						VScroll:  true,
						AssignTo: &h.input,
						Text:     INPUT_DATA,
					},
					TextEdit{
						VScroll:  true,
						AssignTo: &h.resultAll,
					},
				},
			},
			Composite{
				MaxSize: Size{260, 0},
				MinSize: Size{260, 0},
				Layout:  Grid{Columns: 1, MarginsZero: true},
				Children: []Widget{
					Label{
						Text:    "名称",
						MaxSize: Size{0, 15},
					},
					ComboBox{
						AssignTo:      &h.name,
						Editable:      true,
						BindingMember: "Name",
						DisplayMember: "Name",
						MaxSize:       Size{250, 15},
						Model:         data.ReadAllDB(),
						OnCurrentIndexChanged: func() {
							if h.name.CurrentIndex() == -1 {
								return
							}
							data := h.name.Model().([]CliData)[h.name.CurrentIndex()]
							h.bindData(data)
						},
					},
					GroupBox{
						Title:  "执行步骤",
						Layout: Grid{Columns: 5},
						Children: []Widget{
							Label{
								Text:    "从",
								MaxSize: Size{12, 15},
							},
							NumberEdit{
								AssignTo: &h.start,
								MaxSize:  Size{35, 15},
								Value:    0.0,
							},
							Label{
								Text:    "到",
								MaxSize: Size{12, 15},
							},
							NumberEdit{
								AssignTo: &h.end,
								MaxSize:  Size{35, 15},
								Value:    100.0,
							},
						},
					},
					GroupBox{
						Title:  "功能选项",
						Layout: Grid{Columns: 5},
						Children: []Widget{
							CheckBox{
								AssignTo: &h.sendType,
								Text:     "原始上送",
								Checked:  false,
							},
							CheckBox{
								AssignTo: &h.forCheck,
								Text:     "压测",
								Checked:  false,
								OnCheckedChanged: func() {
									if h.forCheck.Checked() {
										h.fTime.SetEnabled(true)
										h.fThread.SetEnabled(true)
									} else {
										h.fTime.SetEnabled(false)
										h.fThread.SetEnabled(false)
									}
								},
							},
							NumberEdit{
								AssignTo: &h.fTime,
								Enabled:  false,
								MaxSize:  Size{35, 15},
								Value:    500.0,
							},
							NumberEdit{
								AssignTo: &h.fThread,
								Enabled:  false,
								MaxSize:  Size{35, 15},
								Value:    200.0,
							},
						},
					},

					Composite{
						Layout: Grid{Columns: 2},
						Children: []Widget{
							PushButton{
								AssignTo: &h.click,
								Text:     "开始",
								OnClicked: func() {
									h.Click()
								},
							},
							PushButton{
								Text: "清空",
								OnClicked: func() {
									h.resultAll.SetText("")
								},
							},
							PushButton{
								AssignTo: &h.click,
								Text:     "保存",
								OnClicked: func() {
									key := h.name.Text()
									data := h.readData()
									data.WriteDB()
									datas := data.ReadAllDB()
									h.name.SetModel(datas)
									for i, d := range datas {
										if d.Name == key {
											h.name.SetCurrentIndex(i)
											break
										}
									}
								},
							},
							PushButton{
								Text: "删除",
								OnClicked: func() {
									index := h.name.CurrentIndex() - 1
									key := h.name.Text()
									data.DelDB(key)
									datas := data.ReadAllDB()
									h.name.SetModel(datas)
									if index > len(datas) {
										index = len(datas)
									} else if index < 0 {
										index = 0
									}
									h.name.SetCurrentIndex(index)
								},
							},
						},
					},
					Label{
						Text:    "二维码Key",
						MaxSize: Size{0, 15},
					},
					LineEdit{
						AssignTo: &h.qrKey,
						Text:     `qr_url":||",`,
					},
					ImageView{
						AssignTo: &h.qrImg,
					},
					PushButton{
						Text: "使用说明",
						OnClicked: func() {
							walk.MsgBox(h.form, "使用说明", INPUT_DATA, walk.MsgBoxOK)
						},
					},
				},
			},
		},
	}.Run()
}

//===========================================http server=========================
type HttpServ struct {
	ServFlag bool
	Form     *walk.MainWindow
	port     *walk.LineEdit
	url      *walk.LineEdit
	read     *walk.TextEdit
	write    *walk.TextEdit
	btn      *walk.PushButton
	data     *ServData
}

func (h *HttpServ) Show() {
	bs, _ := DB.Get([]byte("http.server.data"), nil)
	h.data = new(ServData)
	json.Unmarshal(bs, h.data)

	MainWindow{
		AssignTo: &h.Form,
		Title:    "HTTP服务器",
		MinSize:  Size{400, 300},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{
						Text: "端口",
					},
					LineEdit{
						AssignTo: &h.port,
						Text:     h.data.Port,
						MaxSize:  Size{60, 1},
					},
					Label{
						Text: "路径",
					},
					LineEdit{
						AssignTo: &h.url,
						Text:     h.data.Url,
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
						AssignTo: &h.read,
					},
					Label{
						Text:       "返回",
						ColumnSpan: 2,
					},
					TextEdit{
						AssignTo: &h.write,
						Text:     h.data.Write,
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						AssignTo: &h.btn,
						Text:     "启动",
						OnClicked: func() {
							h.startOrStop()
						},
					},
				},
			},
		},
	}.Run()
}

type ServData struct {
	Port  string
	Url   string
	Read  string
	Write string
}

func (h *HttpServ) startOrStop() {
	h.data.Port = h.port.Text()
	h.data.Url = h.url.Text()
	h.data.Write = h.write.Text()
	if !h.ServFlag {
		if h.data.Url == "" || h.data.Port == "" {
			return
		}
		server := http.NewServeMux()
		server.HandleFunc(h.data.Url, func(w http.ResponseWriter, r *http.Request) {
			ds, _ := ioutil.ReadAll(r.Body)
			h.read.SetText(h.read.Text() + "\r\n" + string(ds) + string(r.URL.Query().Encode()) + string(r.PostForm.Encode()))
			w.Write([]byte(h.write.Text()))
		})
		go manners.ListenAndServe(":"+h.data.Port, server)
		wd, _ := json.Marshal(h.data)
		DB.Put([]byte("http.server.data"), wd, nil)
		h.btn.SetText("停止")
	} else {
		h.btn.SetText("启动")
		manners.Close()
	}
	h.ServFlag = !h.ServFlag
}
