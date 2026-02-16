package generator

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Scanner struct {
	InputDir string
}

func (s *Scanner) scanFile(path string) ([]string, error) {
	results := make([]string, 0)
	f, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "@testgen") {
			results = append(results, line)
		}
	}
	return results, nil
}

func (s *Scanner) getTags(str string) (string, string) {
	str = strings.ReplaceAll(str, "//", "")
	a := strings.SplitSeq(str, " ")
	var endpoint string
	var strct string
	for i := range a {
		if strings.Contains(i, "struct=") {
			strct = strings.ReplaceAll(i, "struct=", "")
		}
		if strings.Contains(i, "router=") {
			endpoint = strings.ReplaceAll(i, "router=", "")
		}

	}
	if endpoint != "" && strct != "" {
		return endpoint, strct
	}
	return "skillIssues", "skillIssues"
}

func (s *Scanner) ScanTags() map[string]map[string]string {
	resultsMap := make(map[string]map[string]string)
	results := make(map[string]string)
	tagList := make([]string, 0)
	filepath.WalkDir(s.InputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		res, err := s.scanFile(path)
		if err != nil {
			log.Println(err)
		}
		tagList = append(tagList, res...)
		return nil
	})
	for _, i := range tagList {
		a, b := s.getTags(i)
		results[a] = b
	}
	for key, item := range results {
		var folder, strct string
		if strings.Contains(item, ".") {
			a := strings.Split(item, ".")
			if len(a) >= 2 {
				folder = "./" + a[0]
				strct = a[1]
			}
		} else {
			folder = "./"
			strct = item

		}
		resultsMap[key] = make(map[string]string)
		resultsMap[key]["folder"] = folder
		resultsMap[key]["struct"] = strct
		resultsMap[key]["name"] = item
	}

	return resultsMap
}
