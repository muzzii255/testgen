package generator

import (
	"bufio"
	"fmt"
	"io/fs"
	"log/slog"
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

func (s *Scanner) getTags(str string) (string, string, error) {
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
		return endpoint, strct, nil
	}
	return "", "", fmt.Errorf("missing router or struct tag in the :%s", str)
}

func (s *Scanner) ScanTags() (map[string]map[string]string, error) {
	resultsMap := make(map[string]map[string]string)
	results := make(map[string]string)
	tagList := make([]string, 0)
	err := filepath.WalkDir(s.InputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("error accessing path", "path", path, "err", err)
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		res, err := s.scanFile(path)
		if err != nil {
			slog.Error("error scanning files", "path", path, "err", err)
			return err
		}
		tagList = append(tagList, res...)
		return nil
	})
	if err != nil {
		return resultsMap, fmt.Errorf("error reading directory %s :%v", s.InputDir, err)
	}
	for _, i := range tagList {
		router, strct, err := s.getTags(i)
		if err != nil {
			slog.Error("error extracting tags from string", "string", i, "err", err)
			continue
		}
		results[router] = strct
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

	return resultsMap, nil
}
