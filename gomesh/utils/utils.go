package utils

import (
	"errors"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func GenPeerId() *uuid.UUID {
	u4, err := uuid.NewV4()
	if err != nil {
		log.Fatalln(err)
	}
	return u4
}

func FormatGuid(b []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func GetNetIP() (interface{}, error) {
	return "112.210.47.38", nil

	res, err := http.Get("http://jsonip.com")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	raw_json, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	json_str := strings.TrimSpace(string(raw_json))
	re := regexp.MustCompile("\"ip\":\"(.+)\",\"about")
	matches := re.FindStringSubmatch(json_str)
	if len(matches) < 2 {
		return nil, errors.New("Unable to find IP match")
	}
	return matches[1], nil
}

func ParseHeaders(h string) map[string]string {
	h_list := strings.Split(strings.TrimSpace(h), "\r\n")
	h_map := make(map[string]string)
	var i_key string
	for _, h := range h_list {
		if string(h[0]) == " " && len(i_key) > 0 {
			if _, ok := h_map[i_key]; ok {
				h_map[i_key] += " " + strings.TrimSpace(h)
			}
		} else if strings.Contains(h, ":") {
			h_frag := strings.Split(h, ": ")
			h_key := h_frag[0]
			h_val := h_frag[1]
			if _, ok := h_map[h_key]; ok {
				h_map[h_key] += "," + h_val
			} else {
				h_map[h_key] = h_val
			}
			i_key = h_key
		} else {
			h_map["Title"] = h
		}
	}
	return h_map
}

func CreateIfNotExist(f_name string) error {
	if _, err := os.Stat(f_name); os.IsNotExist(err) {
		if _, err := os.Create(f_name); err != nil {
			e := fmt.Sprintf("Unable to create file ", f_name, ": ", err)
			return errors.New(e)
		}
	}
	return nil
}
