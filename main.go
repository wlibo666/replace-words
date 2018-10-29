package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ruleFile   = flag.String("rule-file", "", "--rule-file rules.conf")
	targetFile = flag.String("target-file", "", "--target-file target.txt")
	sepChar    = flag.String("sep-char", "blank", "--sep-char blank|tab")
	debug      = flag.Bool("debug", false, "--debug true|false")
)

func getRules(file string) ([][2]string, error) {
	var rules [][2]string

	sep := ' '
	if *sepChar == "tab" {
		sep = '\t'
	}

	fp, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open file:%s failed,err:%s\n", file, err.Error())
		return rules, err
	}
	reader := bufio.NewReader(fp)
	line := ""
	for {
		line, err = reader.ReadString('\n')
		if line != "" {
			line = strings.Trim(line, "\n")
			tmpField := bytes.SplitN([]byte(line), []byte{byte(sep)}, 2)
			if len(tmpField) == 2 {
				rule := [2]string{string(tmpField[0]), string(tmpField[1])}
				rules = append(rules, rule)
			}
		}
		if err == io.EOF {
			break
		}
		line = ""
	}
	return rules, nil
}

func replace(rules [][2]string, file string) (bool, error) {
	if len(rules) == 0 {
		return false, nil
	}

	newFile := file + ".new"
	fileInfo, err := os.Stat(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stat failed,err:%s\n", err.Error())
		return false, err
	}
	fp, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open file:%s failed,err:%s\n", file, err.Error())
		return false, err
	}
	defer fp.Close()
	newFp, err := os.OpenFile(newFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open file:%s failed,err:%s\n", newFile, err.Error())
		return false, err
	}
	defer newFp.Close()

	reader := bufio.NewReader(fp)
	line := ""
	lineNum := 1
	for {
		line, err = reader.ReadString('\n')
		if line != "" {
			oldLine := line
			for _, rule := range rules {
				line = strings.Replace(line, rule[0], rule[1], -1)
			}
			if *debug && oldLine != line {
				fmt.Fprintf(os.Stdout, "line %d changed.\n", lineNum)
				fmt.Fprintf(os.Stdout, "old content:%s", oldLine)
				fmt.Fprintf(os.Stdout, "new content:%s\n", line)
			}
			newFp.WriteString(line)
		}
		if err == io.EOF {
			break
		}
		lineNum++
		line = ""
	}
	newFp.Chmod(fileInfo.Mode())
	return true, nil
}

func rename(file string) error {
	newFile := file + ".new"

	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	_, err = os.Stat(newFile)
	if err != nil {
		return err
	}
	tmpFile := file + ".tmp"
	err = os.Rename(file, tmpFile)
	if err != nil {
		return err
	}
	err = os.Rename(newFile, file)
	if err != nil {
		return err
	}
	err = os.Remove(tmpFile)
	if err != nil {
		return err
	}
	return nil
}

func printRules(rules [][2]string) {
	for index, rule := range rules {
		fmt.Fprintf(os.Stdout, "rule index:%d,src:%s,dst:%s\n", index, rule[0], rule[1])
	}
	fmt.Fprintf(os.Stdout, "\n")
}

func main() {
	flag.Parse()
	if *ruleFile == "" || *targetFile == "" {
		os.Exit(0)
		return
	}
	rules, err := getRules(*ruleFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getRules failed:%s\n", err.Error())
		os.Exit(1)
		return
	}
	if *debug {
		fmt.Fprintf(os.Stdout, "replace rules:\n")
		printRules(rules)
	}
	ok, err := replace(rules, *targetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replace failed:%s\n", err.Error())
		os.Exit(2)
		return
	}
	if ok {
		err = rename(*targetFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rename failed:%s\n", err.Error())
			os.Exit(3)
			return
		}
	}
	os.Exit(0)
}
