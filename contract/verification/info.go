package verification

import (
	"bytes"
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type WasmInfo interface {
	Hash(data []byte) ([]byte, error)
}

type WasmInfoImpl struct {
	url string
}

func NewWasmInfo(url string) WasmInfo {
	return &WasmInfoImpl{url: url}
}

func (wi *WasmInfoImpl) Hash(data []byte) ([]byte, error) {
	httpReq, err := http.NewRequest("POST", wi.url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "text/xml")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("unable to send request, status code: %v", resp.StatusCode))
	}
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	hashHex := strings.TrimSpace("0x" + string(respBody))
	return hexutil.Decode(hashHex)
}
