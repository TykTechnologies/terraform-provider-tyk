package tyk

import (
	"errors"
	"fmt"
	"strings"
)

func AbbreviateDirection(direction string) string {
	switch direction {
	case "east":
		return "e"
	case "north":
		return "n"
	case "south":
		return "s"
	case "northeast":
		return "ne"
	case "northwest":
		return "nw"
	case "west":
		return "w"
	case "southwest":
		return "sw"
	case "southeast":
		return "se"
	case "central":
		return "c"
	}

	return ""
}
func GenerateUrlFromZone(region string, useStaging bool) (string, error) {
	regionPart := strings.Split(region, "-")
	if len(regionPart) != 4 {
		return "", errors.New("the format of this region is wrong")
	}
	suffix := "cloud-ara.tyk.io:37001"
	if useStaging {
		suffix = "ara-staging.tyk.technology:37001"
	}
	url := fmt.Sprintf("https://controller-aws-%s%s%s.%s", regionPart[1], AbbreviateDirection(regionPart[2]), regionPart[3], suffix)
	return url, nil
}
