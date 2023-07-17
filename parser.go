package aini

import (
	"bufio"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

// state enum
type state int

const (
	hostsState    state = 0
	childrenState state = 1
	varsState     state = 2
)

func getState(str string) (state, bool) {
	var result state
	var ok bool = true
	if str == "" || str == "hosts" {
		result = hostsState
	} else if str == "children" {
		result = childrenState
	} else if str == "vars" {
		result = varsState
	} else {
		ok = false
	}
	return result, ok
}

// state enum end

// parser performs parsing of inventory file from some Reader
func (inventory *InventoryData) parse(reader *bufio.Reader) error {
	// This regexp is copy-pasted from ansible sources
	sectionRegex := regexp.MustCompile(`^\[([^:\]\s]+)(?::(\w+))?\]\s*(?:\#.*)?$`)
	scanner := bufio.NewScanner(reader)
	inventory.Groups = make(map[string]*Group)
	inventory.Hosts = make(map[string]*Host)
	activeState := hostsState
	activeGroup := inventory.getOrCreateGroup("ungrouped")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		matches := sectionRegex.FindAllStringSubmatch(line, -1)
		if matches != nil {
			activeGroup = inventory.getOrCreateGroup(matches[0][1])
			var ok bool
			if activeState, ok = getState(matches[0][2]); !ok {
				return fmt.Errorf("section [%s] has unknown type: %s", line, matches[0][2])
			}

			continue
		} else if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			return fmt.Errorf("invalid section entry: '%s'. Make sure that there are no spaces or other characters in the section entry", line)
		}

		if activeState == hostsState {
			hosts, err := inventory.getHosts(line, activeGroup)
			if err != nil {
				return err
			}
			for _, host := range hosts {
				host.DirectGroups[activeGroup.Name] = activeGroup
				inventory.Hosts[host.Name] = host
				if activeGroup.Name != "ungrouped" {
					delete(host.DirectGroups, "ungrouped")
				}
			}
		}
		if activeState == childrenState {
			parsed, err := shlex.Split(line)
			if err != nil {
				return err
			}
			groupName := parsed[0]
			newGroup := inventory.getOrCreateGroup(groupName)
			newGroup.DirectParents[activeGroup.Name] = activeGroup
			inventory.Groups[line] = newGroup
		}
		if activeState == varsState {
			k, v, err := splitKV(line)
			if err != nil {
				return err
			}
			activeGroup.InventoryVars[k] = v
		}
	}
	inventory.Groups[activeGroup.Name] = activeGroup
	return nil
}

// getHosts parses given "host" line from inventory
func (inventory *InventoryData) getHosts(line string, group *Group) (map[string]*Host, error) {
	parts, err := shlex.Split(line)
	if err != nil {
		return nil, err
	}
	hostpattern, port, err := getHostPort(parts[0])
	if err != nil {
		return nil, err
	}
	hostnames, err := expandHostPattern(hostpattern)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Host, len(hostnames))
	for _, hostname := range hostnames {
		params := parts[1:]
		vars := make(map[string]string, len(params))
		for _, param := range params {
			k, v, err := splitKV(param)
			if err != nil {
				return nil, err
			}
			vars[k] = v
		}

		host := inventory.getOrCreateHost(hostname)
		host.Port = port
		host.DirectGroups[group.Name] = group
		addValues(host.InventoryVars, vars)

		result[host.Name] = host
	}
	return result, nil
}

// splitKV splits `key=value` into two string: key and value
func splitKV(kv string) (string, string, error) {
	keyval := strings.SplitN(kv, "=", 2)
	if len(keyval) == 1 {
		return "", "", fmt.Errorf("bad key=value pair supplied: %s", kv)
	}
	return strings.TrimSpace(keyval[0]), strings.TrimSpace(keyval[1]), nil
}

// getHostPort splits string like `host-[a:b]-c:22` into `host-[a:b]-c` and `22`
func getHostPort(str string) (string, int, error) {
	port := 22
	parts := strings.Split(str, ":")
	if len(parts) == 1 {
		return str, port, nil
	}
	lastPart := parts[len(parts)-1]
	if strings.Contains(lastPart, "]") {
		// We are in expand pattern, so no port were specified
		return str, port, nil
	}
	port, err := strconv.Atoi(lastPart)
	return strings.Join(parts[:len(parts)-1], ":"), port, err
}

// expandHostPattern turns `host-[a:b]-c` into a flat list of hosts
func expandHostPattern(hostpattern string) ([]string, error) {
	lbrac := strings.Replace(hostpattern, "[", "|", 1)
	rbrac := strings.Replace(lbrac, "]", "|", 1)
	parts := strings.Split(rbrac, "|")

	if len(parts) == 1 {
		// No pattern detected
		return []string{hostpattern}, nil
	}
	if len(parts) != 3 {
		return nil, fmt.Errorf("wrong host pattern: %s", hostpattern)
	}

	head, nrange, tail := parts[0], parts[1], parts[2]
	bounds := strings.Split(nrange, ":")
	if len(bounds) < 2 || len(bounds) > 3 {
		return nil, fmt.Errorf("wrong host pattern: %s", hostpattern)
	}

	var begin, end []rune
	var step = 1
	if len(bounds) == 3 {
		step, _ = strconv.Atoi(bounds[2])
	}

	end = []rune(bounds[1])
	if bounds[0] == "" {
		if isRunesNumber(end) {
			format := fmt.Sprintf("%%0%dd", len(end))
			begin = []rune(fmt.Sprintf(format, 0))
		} else {
			return nil, fmt.Errorf("skipping range start in not allowed with alphabetical range: %s", hostpattern)
		}
	} else {
		begin = []rune(bounds[0])
	}

	var chars []int
	isNumberRange := false

	if isRunesNumber(begin) && isRunesNumber(end) {
		chars = makeRange(runesToInt(begin), runesToInt(end), step)
		isNumberRange = true
	} else if !isRunesNumber(begin) && !isRunesNumber(end) && len(begin) == 1 && len(end) == 1 {
		dict := append(makeRange('a', 'z', 1), makeRange('A', 'Z', 1)...)
		chars = makeRange(
			find(dict, int(begin[0])),
			find(dict, int(end[0])),
			step,
		)
		for i, c := range chars {
			chars[i] = dict[c]
		}
	}

	if len(chars) == 0 {
		return nil, fmt.Errorf("bad range specified: %s", nrange)
	}

	var hosts []string
	var format string
	if isNumberRange {
		format = fmt.Sprintf("%%s%%0%dd%%s", len(begin))
	} else {
		format = "%s%c%s"
	}

	for _, c := range chars {
		hosts = append(hosts, fmt.Sprintf(format, head, c, tail))
	}

	var result []string
	for _, hostpattern := range hosts {
		newHosts, err := expandHostPattern(hostpattern)
		if err != nil {
			return nil, err
		}
		result = append(result, newHosts...)
	}

	return result, nil
}

func isRunesNumber(runes []rune) bool {
	for _, rune := range runes {
		if rune < '0' || rune > '9' {
			return false
		}
	}
	return true
}

// runesToInt turn runes into corresponding number, ex. '7' -> 7
// should be called only on "number" runes! (see `isRunesNumber` function)
func runesToInt(runes []rune) int {
	result := 0
	for i, rune := range runes {
		result += int((rune - '0')) * int(math.Pow10(len(runes)-1-i))
	}
	return result
}

func makeRange(start, end, step int) []int {
	s := make([]int, 0, 1+(end-start)/step)
	for start <= end {
		s = append(s, start)
		start += step
	}
	return s
}

func find(a []int, x int) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}
