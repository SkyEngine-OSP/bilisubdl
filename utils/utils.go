package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func Request(url string, query map[string]string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for j, s := range query {
		q.Add(j, s)
	}

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %s", resp.Status)
	}

	return resp.Body, nil
}

func JsonUnmarshal(r io.ReadCloser, t interface{}) error {
	err := json.NewDecoder(r).Decode(&t)
	if err != nil {
		return err
	}
	defer r.Close()
	return nil
}

func SecondToTime(tt float64) string {
	secs, msec := int64(tt), int64(tt*1000)%1000
	mins, secs := secs/60, secs%60
	hrs, mins := mins/60, mins%60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hrs, mins, secs, msec)
}

func CleanText(t string) string {
	toBeReplaces := []string{"\"", "?", "/", ":", "\\", "*", "<", ">", "|"}
	for _, elem := range toBeReplaces {
		t = strings.ReplaceAll(t, elem, "_")
	}
	for _, elem := range []string{"\n", "\t"} {
		t = strings.ReplaceAll(t, elem, " ")
	}

	return strings.TrimSpace(strings.TrimRight(t, "."))
}

func WriteFile(filename string, content []byte, mTime time.Time) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(content); err != nil {
		return err
	}

	if err = os.Chtimes(filename, mTime, mTime); err != nil {
		return err
	}

	return nil
}

func ListSelect(list []string, max int) []int {
	var item []int
	for _, s := range list {
		if b := strings.Split(s, "-"); len(b) > 1 {
			b0, _ := strconv.Atoi(b[0])
			b1, _ := strconv.Atoi(b[1])
			for i := b0; i <= b1; i++ {
				if i <= max {
					item = append(item, i)
				}
			}
		} else {
			a, _ := strconv.Atoi(s)
			if a <= max {
				item = append(item, a)
			}
		}
	}
	return item
}
