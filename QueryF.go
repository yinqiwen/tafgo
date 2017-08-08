// **********************************************************************
// This file was generated by a TAF parser!
// TAF version 3.2.2.2 by WSRD Tencent.
// Generated from `QueryF.jce'
// **********************************************************************

package tafgo

import (
	"bytes"
	"time"
)

type QueryF interface {
	FindObjectById(id string, context map[string]string) ([]EndpointF, map[string]string, error)
	FindObjectById4Any(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (int32, map[string]string, error)
	FindObjectById4All(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (int32, map[string]string, error)
	FindObjectByIdInSameGroup(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (int32, map[string]string, error)
	FindObjectByIdInSameStation(id string, sStation string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (int32, map[string]string, error)
	FindObjectByIdInSameSet(id string, setId string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (int32, map[string]string, error)
}

/* proxy for client */
type QueryFProxy struct {
	TafClient *Client
}

func (p *QueryFProxy) FindObjectById(id string, context map[string]string) (_ret []EndpointF, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectById", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagVectorValue(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *QueryFProxy) FindObjectById4Any(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectById4Any", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, activeEp, 2, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, inactiveEp, 3, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *QueryFProxy) FindObjectById4All(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectById4All", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, activeEp, 2, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, inactiveEp, 3, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *QueryFProxy) FindObjectByIdInSameGroup(id string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectByIdInSameGroup", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, activeEp, 2, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, inactiveEp, 3, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *QueryFProxy) FindObjectByIdInSameStation(id string, sStation string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	EncodeTagStringValue(&osBuffer, sStation, 2)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectByIdInSameStation", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, activeEp, 3, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, inactiveEp, 4, true)
	if nil != tafErr {
		return
	}
	return
}
func (p *QueryFProxy) FindObjectByIdInSameSet(id string, setId string, activeEp *[]EndpointF, inactiveEp *[]EndpointF, context map[string]string) (_ret int32, respContext map[string]string, tafErr error) {
	var osBuffer bytes.Buffer
	EncodeTagStringValue(&osBuffer, id, 1)
	EncodeTagStringValue(&osBuffer, setId, 2)
	rep, err := p.TafClient.Invoke(JCENORMAL, "findObjectByIdInSameSet", &osBuffer, context)
	if nil != err {
		tafErr = err
		return
	}

	respContext = rep.Context
	respBuffer := bytes.NewBuffer(rep.SBuffer)
	tafErr = DecodeTagInt32Value(respBuffer, &_ret, 0, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, activeEp, 3, true)
	if nil != tafErr {
		return
	}
	tafErr = DecodeTagVectorValue(respBuffer, inactiveEp, 4, true)
	if nil != tafErr {
		return
	}
	return
}

func NewQueryFProxy(obj string, timeout time.Duration) *QueryFProxy {
	c := NewClient(obj, timeout)
	proxy := &QueryFProxy{c}
	return proxy
}