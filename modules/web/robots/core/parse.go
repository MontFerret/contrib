package core

import (
	"bufio"
	"strconv"
	"strings"
)

const maxScanTokenSize = 1024 * 1024

// Parse decodes raw robots.txt text into a plain document shape.
func Parse(text string) (Document, error) {
	doc := Document{
		Groups:   make([]Group, 0),
		Sitemaps: make([]string, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)

	var current *Group
	var currentSealed bool
	lineNo := 0

	for scanner.Scan() {
		lineNo++

		name, value, ok := parseLine(scanner.Text())
		if !ok {
			continue
		}

		switch strings.ToLower(name) {
		case "user-agent":
			if value == "" {
				return Document{}, newErrorf(StageParse, "line %d: user-agent must not be empty", lineNo)
			}

			if current == nil || currentSealed {
				doc.Groups = append(doc.Groups, Group{
					UserAgents: make([]string, 0, 1),
					Allow:      make([]string, 0),
					Disallow:   make([]string, 0),
				})
				current = &doc.Groups[len(doc.Groups)-1]
				currentSealed = false
			}

			current.UserAgents = append(current.UserAgents, value)
		case "allow":
			if current == nil || len(current.UserAgents) == 0 {
				continue
			}

			current.Allow = append(current.Allow, value)
			currentSealed = true
		case "disallow":
			if current == nil || len(current.UserAgents) == 0 {
				continue
			}

			current.Disallow = append(current.Disallow, value)
			currentSealed = true
		case "crawl-delay":
			if current == nil || len(current.UserAgents) == 0 {
				continue
			}

			if value == "" {
				return Document{}, newErrorf(StageParse, "line %d: crawl-delay must not be empty", lineNo)
			}

			delay, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return Document{}, newErrorf(StageParse, "line %d: crawl-delay must be numeric", lineNo)
			}

			current.CrawlDelay = &delay
			currentSealed = true
		case "sitemap":
			if value != "" {
				doc.Sitemaps = append(doc.Sitemaps, value)
			}
		case "host":
			if value != "" {
				host := value
				doc.Host = &host
			}
		default:
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return Document{}, wrapError(StageParse, err, "failed to scan robots document")
	}

	return doc, nil
}

func parseLine(line string) (name, value string, ok bool) {
	if idx := strings.IndexByte(line, '#'); idx >= 0 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}

	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return "", "", false
	}

	name = strings.TrimSpace(line[:idx])
	if name == "" {
		return "", "", false
	}

	return name, strings.TrimSpace(line[idx+1:]), true
}
