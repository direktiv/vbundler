package calver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidCalVer = errors.New("invalid version string")

type CalVer string

func Parse(s string) (CalVer, error) {

	k := strings.Count(s, ".")
	if k < 1 || k > 2 {
		return CalVer(""), ErrInvalidCalVer
	}

	k = strings.Count(s, "-")
	if k > 1 {
		return CalVer(""), ErrInvalidCalVer
	}

	if idx := strings.Index(s, "-"); idx != -1 && strings.LastIndex(s, ".") > idx {
		return CalVer(""), ErrInvalidCalVer
	}

	elems := strings.Split(s, "-")
	elems = strings.Split(elems[0], ".")
	for _, elem := range elems {
		_, err := strconv.Atoi(elem)
		if err != nil {
			return CalVer(""), ErrInvalidCalVer
		}
	}

	v := CalVer(s)

	if v.Modifier() != "" && v.Patch() == -1 {
		return CalVer(""), ErrInvalidCalVer
	}

	return v, nil
}

func (v CalVer) String() string {
	return string(v)
}

func (v CalVer) MarshalText() (text []byte, err error) {
	return []byte(v.String()), nil
}

func (v *CalVer) UnmarshalText(text []byte) error {
	var err error
	*v, err = Parse(string(text))
	if err != nil {
		return err
	}
	return nil
}

func (v CalVer) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *CalVer) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, "\"")
	var err error
	*v, err = Parse(s)
	if err != nil {
		return err
	}
	return nil
}

func (v CalVer) Year() int {
	elems := strings.SplitN(string(v), ".", 2)
	x, err := strconv.Atoi(elems[0])
	if err != nil {
		panic(ErrInvalidCalVer)
	}
	return x
}

func (v CalVer) Major() int {
	return v.Year()
}

func (v CalVer) Month() int {
	elems := strings.SplitN(string(v), ".", 2)
	if len(elems) < 2 {
		panic(ErrInvalidCalVer)
	}
	idx := strings.IndexAny(elems[1], ".-")
	var s string
	if idx == -1 {
		s = elems[1]
	} else {
		s = elems[1][:idx]
	}
	x, err := strconv.Atoi(s)
	if err != nil {
		panic(ErrInvalidCalVer)
	}
	return x
}

func (v CalVer) Minor() int {
	return v.Month()
}

func (v CalVer) Patch() int {
	elems := strings.SplitN(string(v), ".", 3)
	if len(elems) < 3 {
		return -1
	}
	elems = strings.SplitN(string(v), "-", 2)
	elems = strings.Split(elems[0], ".")
	x, err := strconv.Atoi(elems[2])
	if err != nil {
		panic(ErrInvalidCalVer)
	}
	return x
}

func (v CalVer) Modifier() string {
	elems := strings.SplitN(string(v), "-", 2)
	if len(elems) < 2 {
		return ""
	}
	return elems[1]
}

func (v CalVer) Less(version CalVer) bool {
	if v.Major() <= version.Major() && v.Minor() <= version.Minor() &&
		(version.Patch() == -1 || (v.Patch() != -1 && v.Patch() <= version.Patch())) {

		if v.Major() == version.Major() && version.Minor() == version.Minor() &&
			v.Patch() == version.Patch() {

			if v.Modifier() == "" {
				return false
			}

			if version.Modifier() == "" {
				return true
			}

			return v.Modifier() < version.Modifier()
		}

		return true
	}
	return false
}

type CalVers []CalVer

func (a CalVers) Len() int {
	return len(a)
}

func (a CalVers) Less(i, j int) bool {
	return a[i].Less(a[j])
}

func (a CalVers) Swap(i, j int) {
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}

func (a CalVers) BestMatch(v CalVer) (CalVer, error) {

	idx := sort.Search(a.Len(), func(arg1 int) bool {
		return v.Less(a[arg1])
	})
	if idx < a.Len() && a[idx].String() == v.String() {
		return v, nil // exact match
	}

	// if v has a modifier it demands an exact match
	if v.Modifier() != "" {
		return v, fmt.Errorf("no match for kernel %s", v.String())
	}

	if idx != 0 {
		candidate := a[idx-1]
		if candidate.Major() == v.Major() && candidate.Minor() == candidate.Minor() {
			if v.Patch() == -1 || candidate.Patch() == v.Patch() {
				return candidate, nil
			}
		}
	}
	return v, fmt.Errorf("no match for kernel %s", v.String())
}
