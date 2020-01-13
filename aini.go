package aini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

// InventoryData contains parsed inventory representation
type InventoryData struct {
	Groups map[string]*Group
	Hosts  map[string]*Host
}

// Group represents ansible group
type Group struct {
	Name     string
	Vars     map[string]string
	Hosts    map[string]*Host
	Children map[string]*Group
	Parents  map[string]*Group
}

// Host represents ansible host
type Host struct {
	Name   string
	Port   int
	Vars   map[string]string
	Groups map[string]*Group
}

// ParseFile parses the file
func ParseFile(f string) (*InventoryData, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &InventoryData{}, err
	}

	inventory, err := Parse(bytes.NewReader(bs))
	if err != nil {
		return &InventoryData{}, err
	}

	return inventory, nil
}

// Parse using some Reader
func Parse(r io.Reader) (*InventoryData, error) {
	input := bufio.NewReader(r)
	inventory := &InventoryData{}
	inventory.parse(input)
	inventory.Reconcile()
	return inventory, nil
}

// Reconcile ensures inventory basic rules, run after updates
func (inventory *InventoryData) Reconcile() {
	allGroup := inventory.getGroup("all")
	allGroup.Hosts = inventory.Hosts
	allGroup.Children = inventory.Groups

	for _, host := range inventory.Hosts {
		for _, group := range host.Groups {
			ancestors := group.getAncestors()
			host.setVarsIfNotExist(group.Vars)
			for _, ancestor := range ancestors {
				ancestor.Hosts[host.Name] = host
				ancestor.Children[group.Name] = group
				for k, v := range ancestor.Vars {
					if _, ok := host.Vars[k]; !ok {
						host.Vars[k] = v
					}
					if _, ok := group.Vars[k]; !ok {
						group.Vars[k] = v
					}
				}
			}
		}
		host.setVarsIfNotExist(allGroup.Vars)
		host.Groups["all"] = allGroup
	}
	inventory.Groups["all"] = allGroup
}

// Match looks for a hosts that match the pattern
func (inventory *InventoryData) Match(m string) []*Host {
	matchedHosts := make([]*Host, 0)
	for _, host := range inventory.Hosts {
		if m, err := path.Match(m, host.Name); err == nil && m {
			matchedHosts = append(matchedHosts, host)
		}
	}
	return matchedHosts
}

type state int

const (
	hostsState    state = 0
	childrenState state = 1
	varsState     state = 2
)

func (inventory *InventoryData) parse(reader *bufio.Reader) error {
	// This regex is copy-pasted from ansible sources
	sectionRegex := regexp.MustCompile(`^\[([^:\]\s]+)(?::(\w+))?\]\s*(?:\#.*)?$`)
	scanner := bufio.NewScanner(reader)
	inventory.Groups = make(map[string]*Group)
	inventory.Hosts = make(map[string]*Host)
	activeState := hostsState
	activeGroup := inventory.getGroup("ungrouped")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		matches := sectionRegex.FindAllStringSubmatch(line, -1)
		if matches != nil {
			inventory.Groups[activeGroup.Name] = activeGroup
			activeGroup = inventory.getGroup(matches[0][1])
			var ok bool
			if activeState, ok = getState(matches[0][2]); !ok {
				return fmt.Errorf("Section [%s] has unknown type: %s", line, matches[0][2])
			}

			continue
		} else if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			return fmt.Errorf("Invalid section entry: '%s'. Please make sure that there are no spaces in the section entry, and that there are no other invalid characters", line)
		}

		if activeState == hostsState {
			hosts, err := getHosts(line, activeGroup)
			if err != nil {
				return err
			}
			for k, v := range hosts {
				activeGroup.Hosts[k] = v
			}
			for _, host := range hosts {
				inventory.Hosts[host.Name] = host
			}
		}
		if activeState == childrenState {
			newGroup := inventory.getGroup(line)
			newGroup.Parents[activeGroup.Name] = activeGroup
			inventory.Groups[line] = newGroup
		}
		if activeState == varsState {
			k, v, err := splitKV(line)
			if err != nil {
				return err
			}
			activeGroup.Vars[k] = v
		}
	}
	inventory.Groups[activeGroup.Name] = activeGroup
	return nil
}

func (inventory *InventoryData) getGroup(groupName string) *Group {
	if group, ok := inventory.Groups[groupName]; ok {
		return group
	}
	return &Group{
		Name:     groupName,
		Hosts:    make(map[string]*Host, 0),
		Vars:     make(map[string]string, 0),
		Children: make(map[string]*Group, 0),
		Parents:  make(map[string]*Group, 0),
	}

}

// splitKV splits `key=value` into two string: key and value
func splitKV(kv string) (string, string, error) {
	parsed, err := shlex.Split(kv)
	if err != nil {
		return "", "", err
	}

	keyval := strings.Split(strings.Join(parsed, ""), "=")
	if len(keyval) != 2 {
		return "", "", fmt.Errorf("Bad key=value pair supplied: %s", kv)
	}
	return strings.TrimSpace(keyval[0]), strings.TrimSpace(keyval[1]), nil
}

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

func getHosts(line string, group *Group) (map[string]*Host, error) {
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

		host := &Host{Name: hostname, Port: port, Vars: vars, Groups: map[string]*Group{group.Name: group}}
		result[host.Name] = host
	}
	return result, nil
}

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

func expandHostPattern(hostpattern string) ([]string, error) {
	lbrac := strings.Replace(hostpattern, "[", "|", 1)
	rbrac := strings.Replace(lbrac, "]", "|", 1)
	parts := strings.Split(rbrac, "|")

	if len(parts) == 1 {
		// No pattern detected
		return []string{hostpattern}, nil
	}
	if len(parts) != 3 {
		return nil, fmt.Errorf("Wrong host pattern: %s", hostpattern)
	}

	head, nrange, tail := parts[0], parts[1], parts[2]
	bounds := strings.Split(nrange, ":")
	if len(bounds) < 2 || len(bounds) > 3 {
		return nil, fmt.Errorf("Wrong host pattern: %s", hostpattern)
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
			return nil, fmt.Errorf("Skipping range start in not allowed with alphabetical range: %s", hostpattern)
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
		return nil, fmt.Errorf("Bad range specified: %s", nrange)
	}

	var result []string
	var format string
	if isNumberRange {
		format = fmt.Sprintf("%%s%%0%dd%%s", len(begin))
	} else {
		format = "%s%c%s"
	}

	for _, c := range chars {
		result = append(result, fmt.Sprintf(format, head, c, tail))
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

func (host *Host) setVarsIfNotExist(vars map[string]string) {
	for k, v := range vars {
		if _, ok := host.Vars[k]; !ok {
			host.Vars[k] = v
		}
	}
}

// getAncestors returns all Ancestors of a given group in level order
func (group *Group) getAncestors() []*Group {
	result := make([]*Group, 0)

	for queue := []*Group{group}; ; {
		group := queue[0]
		parentList := mapValuesList(group.Parents)
		result = append(result, parentList...)
		copy(queue, queue[1:])
		queue = queue[:len(queue)-1]
		queue = append(queue, parentList...)

		if len(queue) == 0 {
			return result
		}
	}
}

func mapValuesList(mymap map[string]*Group) []*Group {
	values := make([]*Group, len(mymap))

	i := 0
	for _, v := range mymap {
		values[i] = v
		i++
	}
	return values
}
