package myacc2

import (
	"fmt"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// set elements 0 through n-1 to c
func fill(v []int, n, c int) {
	for i := 0; i < n; i++ {
		v[i] = c
	}
}

func extend[T any](s *[]T, INC int) {
	new := make([]T, len(*s)+INC)
	copy(new, *s)
	*s = new
}

func printWset() {
	buf := strings.Builder{}
	buf.WriteString("[]wItem{\n")
	for _, e := range wSet {
		if e.item.prd.prd != nil {
			buf.WriteString("{item{\n")
			buf.WriteString(fmt.Sprintf("%d,\n", e.item.off))
			buf.WriteString("prd{\n")
			buf.WriteString(fmt.Sprintf("id: %d,\n", e.item.prd.id))
			buf.WriteString(fmt.Sprintf("prd:[]int%s,\n", printSlice(e.item.prd.prd)))
			buf.WriteString("},\n")
			buf.WriteString(fmt.Sprintf("%d,\n", e.item.first))
			buf.WriteString(fmt.Sprintf("lkset%s,\n", printSlice(e.item.lkset)))
			buf.WriteString(fmt.Sprintf("},%t,%t},\n", e.processed, e.done))
		}
	}
	buf.WriteString("}\n")
	fmt.Println(buf.String())
}

func printSlice(s []int) string {
	buf := strings.Builder{}
	buf.WriteRune('{')
	for _, n := range s {
		buf.WriteString(fmt.Sprintf("%d,", n))
	}
	res := buf.String()
	res = res[:len(res)-1] + "}"
	return res
}

func printState() {
	buf := strings.Builder{}
	buf.WriteString("[]item{\n")
	for _, a := range kernls {
		if a.prd.prd == nil {
			break
		}
		buf.WriteString(fmt.Sprintf("{%d,prd{%d,[]int%s},%d,lkset%s},\n", a.off, a.prd.id, printSlice(a.prd.prd), a.first, printSlice(a.lkset)))
	}
	buf.WriteString("}\n")
	fmt.Println(buf.String())
}

func printItem(a item) {
	fmt.Printf("===== item: %v, %d->%d\n", a.prd.prd, a.off, a.first)
}

func printItems(s []item) {
	for _, a := range s {
		printItem(a)
	}
}

func isdigit(c rune) bool { return c >= '0' && c <= '9' }

func isword(c rune) bool {
	return c >= 0xa0 || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
