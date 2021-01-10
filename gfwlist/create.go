package gfwlist

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

func CreateGFWListFile(remoteUrl, file string) error {
	var content []byte
	var err error
	var retryCount int
	for {
		content, err = fetchGFWListFromRemote(remoteUrl)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return err
			}
			fmt.Println("get fail,retry...")
			continue
		}
		break
	}

	body, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, []byte(body), 0666)
	if err != nil {
		return err
	}
	return nil
}

func fetchGFWListFromRemote(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid response:%v", resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return nil, err
	}

	return body, nil
}
