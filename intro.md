# tafgo
TAF是MIG后台通用服务框架，但目前并无golang支持，tafgo则是一个针对TAF框架的纯golang支持实现，以下介绍tafgo的实现原理和相关应用示例， 用户可以基于此编写纯golang的TAF client程序。

# 编解码
## 基本类型
基本类型包括 
- int8/int16/int32/int64
- float32/float64
- string

相关实现都是直接参考taf C++源码直接翻译而来，并无难度，这里不再展开详述。 

## 容器类型
由于golang中并无泛型的概念，所以只能基于反射来实现相关编解码。目前支持的容器类型为jce2cpp中的支持的两个:  
- vector
- map

所有编解码实现里，这里是唯一的稍有难度的地方，相关实现可以参考后面代码实现。

## Struct类型
针对Struct的生成规则如下：  
- 通过golang中的struct field tag携带 tag， required， json name相关信息
- 生成Encode/Decode方法
- 针对成员类型，基本类型调用tafgo中的对应Encode/Decode实现；
- map/vector调用tafgo中的对应map/vector的Encode/Decode实现；
- Struct类型则直接调用其Encode/Decode方法

定义一个接口TafStruct， 所有生成的代码遵循此接口约定:   
```go
type TafStruct interface {
	Encode(buf *bytes.Buffer) error
	Decode(buf *bytes.Buffer) error
}
```

理论上代码生成如果更简洁些是可以完全基于反射实现所有数据的Encode/Decode，但考虑性能因素， 目前的所有的编解码实现只有vector/map相关类型无可避免的用到了反射实现。


# 代码生成
参考jce2cpp的代码， 基本只保留了最通用部分的实现，像wup，msf相关的都去掉了。由于TAF C++/Java服务端内置有更复杂的功能实现，如统计，染色，监控上报等等，生成server端代码目前意义不大，所以也去除了server端的代码生成。 
目前代码生成具备如下能力：   
- Struct结构的编解码（忽略了默认值）
- Interface接口的Client端Proxy实现的生成
- 生成相关Struct tag方便周边工具实现，如到json的转换是通过生成的Struct tag与标准json库实现的

struct的代码生成一个简单例子：
```go
type SuggestResultItem struct {
	Suggestion   string `tag:"0"  required:"false"  json:"suggestion"`
	Recommend_id []byte `tag:"1"  required:"false"  json:"recommend_id"`
}

func (p *SuggestResultItem) ClassName() string {
	return "yybbd.SuggestResultItem"
}
func (p *SuggestResultItem) MD5() string {
	return "ba4306b5d9c507ba7dddf1d1fdbbab58"
}
func (p *SuggestResultItem) ResetDefautlt() {
	var empty SuggestResultItem
	*p = empty
}
func (p *SuggestResultItem) Encode(buf *bytes.Buffer) error {
	var err error
	err = taf.EncodeTagStringValue(buf, p.Suggestion, 0)
	if nil != err {
		return err
	}
	err = taf.EncodeTagBytesValue(buf, p.Recommend_id, 1)
	if nil != err {
		return err
	}
	return nil
}
func (p *SuggestResultItem) Decode(buf *bytes.Buffer) error {
	var err error
	err = taf.DecodeTagStringValue(buf, &p.Suggestion, 0, false)
	if nil != err {
		return err
	}
	err = taf.DecodeTagBytesValue(buf, &p.Recommend_id, 1, false)
	if nil != err {
		return err
	}
	return err
}
```

interface代码生成的例子：  
```go
type BDSearchService interface {
	Search(stHead *MobileAssist.BusinessRequestHead, req *BDSearchRequest, rsp *BDSearchResponse, context map[string]string) (int32, map[string]string, error)
	GetSuggestion(stHead *MobileAssist.BusinessRequestHead, req *GetSuggestionRequest, rsp *GetSuggestionResponse, context map[string]string) (int32, map[string]string, error)
	XFSearch(stHead *MobileAssist.BusinessRequestHead, req *XFSearchRequest, rsp *XFSearchResponse, context map[string]string) (int32, map[string]string, error)
}

/* proxy for client */
type BDSearchServiceProxy struct {
	TafClient *taf.Client
}

func (p *BDSearchServiceProxy) Search(stHead *MobileAssist.BusinessRequestHead, req *BDSearchRequest, rsp *BDSearchResponse, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	taf.EncodeTagStructValue(&osBuffer, stHead, 1)
	taf.EncodeTagStructValue(&osBuffer, req, 2)
	rep, err := p.TafClient.Invoke(taf.JCENORMAL, "Search", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = taf.DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = taf.DecodeTagStructValue(respBuffer, rsp, 3, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *BDSearchServiceProxy) GetSuggestion(stHead *MobileAssist.BusinessRequestHead, req *GetSuggestionRequest, rsp *GetSuggestionResponse, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	taf.EncodeTagStructValue(&osBuffer, stHead, 1)
	taf.EncodeTagStructValue(&osBuffer, req, 2)
	rep, err := p.TafClient.Invoke(taf.JCENORMAL, "GetSuggestion", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = taf.DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = taf.DecodeTagStructValue(respBuffer, rsp, 3, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *BDSearchServiceProxy) XFSearch(stHead *MobileAssist.BusinessRequestHead, req *XFSearchRequest, rsp *XFSearchResponse, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	taf.EncodeTagStructValue(&osBuffer, stHead, 1)
	taf.EncodeTagStructValue(&osBuffer, req, 2)
	rep, err := p.TafClient.Invoke(taf.JCENORMAL, "XFSearch", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = taf.DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = taf.DecodeTagStructValue(respBuffer, rsp, 3, true)
	if nil != tafErr {
		return
	}
	return
}

func NewBDSearchServiceProxy(obj string, timeout time.Duration) *BDSearchServiceProxy {
	c := taf.NewClient(obj, timeout)
	proxy := &BDSearchServiceProxy{c}
	return proxy
}
```


# RPC通信
taf的RPC通信机制也很简单，简单描述如下：  
- [Proxy]将RPC方法调用打包到一个[]byte
- [RPC框架]封装一个RequestPacket的jce结构对象
- [RPC框架]发送加上4字节网络字节序的长度头给server
- [RPC框架]接收4字节网络字节序的长度头的ResponsePacket的jce结构对象
- [Proxy]解包ResponsePacket中的sBuffer为RCP调用结果

以下是生成的RequestPacket和ResponsePacket定义：  
```go
type RequestPacket struct {
	IVersion     int16             `tag:"1"  required:"true"`
	CPacketType  byte              `tag:"2"  required:"true"`
	IMessageType int32             `tag:"3"  required:"true"`
	IRequestId   int32             `tag:"4"  required:"true"`
	SServantName string            `tag:"5"  required:"true"`
	SFuncName    string            `tag:"6"  required:"true"`
	SBuffer      []byte            `tag:"7"  required:"true"`
	ITimeout     int32             `tag:"8"  required:"true"`
	Context      map[string]string `tag:"9"  required:"true"`
	Status       map[string]string `tag:"10"  required:"true"`
}


type ResponsePacket struct {
	IVersion     int16             `tag:"1"  required:"true"`
	CPacketType  byte              `tag:"2"  required:"true"`
	IRequestId   int32             `tag:"3"  required:"true"`
	IMessageType int32             `tag:"4"  required:"true"`
	IRet         int32             `tag:"5"  required:"true"`
	SBuffer      []byte            `tag:"6"  required:"true"`
	Status       map[string]string `tag:"7"  required:"true"`
	SResultDesc  string            `tag:"8"  required:"false"`
	Context      map[string]string `tag:"9"  required:"false"`
}
```
由于golang的网络编程相比C++、Java简洁的多，只需要很少几行代码就可实现上述逻辑，如下：  
```go
func (c *Client) Invoke(ctype uint8, funcName string, req *bytes.Buffer, ctx map[string]string) (*ResponsePacket, error) {
	packet := RequestPacket{}
	packet.SBuffer = req.Bytes()
	packet.IVersion = 1
	packet.SServantName = c.servant
	packet.SFuncName = funcName
	packet.IRequestId = atomic.AddInt32(&c.sid, 1)
	packet.Context = ctx
	packet.IMessageType = int32(ctype)
	packet.ITimeout = 1000
	session := c.newRPCSession(packet.IRequestId)
	rpcConn := c.getRPCChannel()
	rpcConn.ch <- &packet
	var err error
	var resp *ResponsePacket
	select {
	case resp = <-session.ch:
		break
	case <-time.After(c.Timeout):
		err = ErrTafRPCTimeout
	}

	if nil == resp && nil == err {
		err = fmt.Errorf("No response recevied, maybe timeout")
	}
	c.closeRPCSession(packet.IRequestId)
	return resp, err
}
```



# 相关实现
代码实现包括两部分，[jce2go](http://git.code.oa.com/qiyingwang/jce2go)与[tafgo](https://github.com/yinqiwen/tafgo)。

## jce2go
[jce2go](http://git.code.oa.com/qiyingwang/jce2go) 放在了内部的git.oa上，C++实现，编译依赖taf框架，直接make编译。编译前确认makefile中的taf路径即可：  
```Makefile
...
TAF_PATH := /usr/local/taf
INCLUDE   += -I${TAF_PATH}/include   -I./
LIB       += -L${TAF_PATH}/lib -lparse -lutil
...
```


## tafgo
[tafgo](https://github.com/yinqiwen/tafgo) 主要为基础编解码实现与RCP通信层，纯go实现， 无其他库依赖。


# Usage & Example
## 编写jce
无特殊要求
## Get jce2go
```shell
   git clone http://git.code.oa.com/qiyingwang/jce2go.git  
   cd jce2go; make
```
## 代码生成
```shell
   jce2go --dir=$GOPATH/src  my.jce
```
注意jce2go最后一步会调用gofmt格式化生成的代码，所以需要提前设置好go的相关环境变量。
## Get tafgo
```shell
   go get -u -v github.com/yinqiwen/tafgo
```
## Client代码
```go
package main

import (
	"MobileAssist"
	"encoding/json"
	"log"
	"os"
	"time"
	"yybbd"

	taf "github.com/yinqiwen/tafgo"
)

func toJson(d interface{}) string {
	b, _ := json.MarshalIndent(d, "", "\t")
	return string(b)
}

func main() {
	taf.NewDefaultNaming("Docker.DockerRegistry.shQueryObj@tcp -h 10.150.163.220  -p 9903 -t 50000", 1*time.Second)
	v, _, err := taf.DefaultNamingService.FindObjectById("MobileAssist.XXXXX.YYYYOBJ", nil)
	log.Printf("####%v  %v", err, toJson(v))
	var active, inactive []taf.EndpointF
	v1, _, err := taf.DefaultNamingService.FindObjectById4All("MobileAssist.XXXXX.YYYYOBJ", &active, &inactive, nil)
	log.Printf("####%v  %v %v %v", err, v1, toJson(active), toJson(inactive))

	//proxy := yybbd.NewBDSearchServiceProxy("MobileAssist.XXXXX.YYYYOBJ@tcp -h 1.1.1.1 -p 10116 -t 60000", 1*time.Second)
	proxy := yybbd.NewBDSearchServiceProxy("MobileAssist.XXXXX.YYYYOBJ", 1*time.Second)
	head := &MobileAssist.BusinessRequestHead{}
	req := &yybbd.BDSearchRequest{}
	res := yybbd.BDSearchResponse{}
	req.Content = "qq"
	if len(os.Args) > 1 {
		req.Content = os.Args[1]
	}
	ret, _, err := proxy.Search(head, req, &res, nil)
	log.Printf("####%v %d %v", err, ret, toJson(&res))
}
```
## Build
在client代码路径下编译：  
```shell
  go build -v .
```
在windows可以编译到linux目标： 
```shell
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v .
```