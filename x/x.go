package x

import (
  "errors"
  "fmt"
  "regexp"
  "strconv"
)

var reg = regexp.MustCompile("\\s*([a-zA-Z0-9\\-._]+):([1-9][0-9]+)\\s*")

type HostAndPort struct {
  Host string
  Port int
}

func (p *HostAndPort) Parse(s string) error {
  sm := reg.FindSubmatch([]byte(s))
  if len(sm) != 3 {
    return errors.New(fmt.Sprintf("invalid host and port: %s.", s))
  }
  p.Host = string(sm[1])
  if num, err := strconv.Atoi(string(sm[2])); err != nil {
    return err
  } else {
    p.Port = num
    return nil
  }
}
