package flags

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/multiformats/go-multiaddr"
)

func formatArrayForOutput(array []string) string {
	return "(" + strings.Join(array, ",") + ")"
}

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return formatArrayForOutput(*i)
}

func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type BootnodeFlags []string

func (i *BootnodeFlags) String() string {
	return formatArrayForOutput(*i)
}

func (i *BootnodeFlags) Set(value string) error {
	if _, err := multiaddr.NewMultiaddr(value); err != nil {
		return err
	}
	*i = append(*i, value)
	return nil
}

func MultiAddrFromDns(addr string, port int) (multiaddr.Multiaddr, error) {
	var version string
	var domain string
	match, err := regexp.MatchString("^/?(dns)(4|6)?/[^-|^/][A-Za-z0-9-]([^-|^/]?)+([\\-\\.]{1}[a-z0-9]+)*\\.[A-Za-z]{2,6}(/?)$", addr)
	if err != nil || !match {
		return nil, errors.New("Invalid DNS address")
	}

	s := strings.Trim(addr, "/")
	split := strings.Split(s, "/")
	if len(split) != 2 {
		return nil, errors.New("Invalid DNS address")
	}
	switch split[0] {
	case "dns":
		version = "dns"
	case "dns4":
		version = "dns4"
	case "dns6":
		version = "dns6"
	default:
		return nil, errors.New("Invalid DNS version")
	}
	domain = split[1]

	multiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/%s/%s/tcp/%d", version, domain, port))
	if err != nil {
		return nil, errors.New("Could not create a multi address")
	}
	return multiAddr, nil
}
