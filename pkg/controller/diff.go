package controller

import (
	"fmt"
	"os"
	"strings"
)

func printDiff(filePath, expected, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	edits := computeDiff(expectedLines, actualLines)
	for _, e := range edits {
		switch e.op {
		case diffDelete:
			fmt.Fprintf(os.Stderr, "%s:%d\n", filePath, e.line)
			fmt.Fprintf(os.Stderr, "-%s\n", e.text)
		case diffInsert:
			fmt.Fprintf(os.Stderr, "%s:%d\n", filePath, e.line)
			fmt.Fprintf(os.Stderr, "+%s\n", e.text)
		}
	}
}

type diffOp int

const (
	diffDelete diffOp = iota
	diffInsert
)

type diffEdit struct {
	op   diffOp
	line int // 1-based line number in expected (delete) or actual (insert)
	text string
}

func computeDiff(expected, actual []string) []diffEdit { //nolint:cyclop
	n, m := len(expected), len(actual)
	// LCS via dynamic programming
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := range n {
		for j := range m {
			if expected[i] == actual[j] {
				dp[i+1][j+1] = dp[i][j] + 1
			} else {
				dp[i+1][j+1] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	// Backtrack to find edits
	var edits []diffEdit
	i, j := n, m
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && expected[i-1] == actual[j-1]:
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			edits = append(edits, diffEdit{op: diffInsert, line: j, text: actual[j-1]})
			j--
		default:
			edits = append(edits, diffEdit{op: diffDelete, line: i, text: expected[i-1]})
			i--
		}
	}
	// Reverse to get correct order
	for l, r := 0, len(edits)-1; l < r; l, r = l+1, r-1 {
		edits[l], edits[r] = edits[r], edits[l]
	}
	return edits
}
