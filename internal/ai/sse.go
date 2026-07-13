package ai

import (
	"bufio"
	"io"
	"strings"
)

func scanSSE(reader io.Reader, handle func(event, data string) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var event string
	var data []string
	dispatch := func() error {
		if len(data) == 0 {
			return nil
		}
		err := handle(event, strings.Join(data, "\n"))
		event = ""
		data = nil
		return err
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data = append(data, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return dispatch()
}
