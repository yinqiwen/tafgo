package tafgo

import (
	"bytes"
	"testing"
)

func TestCodec(t *testing.T) {
	v1 := &RequestPacket{}
	v1.SFuncName = "helloww"
	v1.IMessageType = 12456
	v1.ITimeout = 10101
	v1.SServantName = "343242342$$"
	v1.Context = make(map[string]string)
	v1.Context["AAA"] = "BBB"
	v1.SBuffer = []byte("#######")
	var buf bytes.Buffer
	err := v1.Encode(&buf)
	if nil != err {
		t.Fatalf("###%v", err)
	}
	t.Logf("####%v", buf.Len())

	v2 := &RequestPacket{}
	err = v2.Decode(&buf)
	if nil != err {
		t.Fatalf("###%v", err)
	}
	t.Logf("####%v", v2)
}
