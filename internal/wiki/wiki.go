package wiki

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Article struct {
	Title  string
	Text   []string
	Words  int
	AvgLen float64
}

func Fetch(title string) (*Article, error) {
	escaped := url.PathEscape(title)
	u := fmt.Sprintf("https://fr.wikipedia.org/wiki/%s", escaped)

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FileOpsBot/1.0 (+https://example.com/contact)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d pour %s", resp.StatusCode, u)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var paras []string
	doc.Find("p").Each(func(_ int, sel *goquery.Selection) {
		if txt := strings.TrimSpace(sel.Text()); txt != "" {
			paras = append(paras, txt)
		}
	})

	words, avg := statsWords(paras)

	return &Article{
		Title:  title,
		Text:   paras,
		Words:  words,
		AvgLen: avg,
	}, nil
}

func Save(a *Article, outDir string) (string, error) {
	name := fmt.Sprintf("wiki_%s.txt", a.Title)
	path := filepath.Join(outDir, name)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== %s ===\n\n", a.Title))
	for _, p := range a.Text {
		b.WriteString(p + "\n\n")
	}
	b.WriteString(fmt.Sprintf("--- Stats ---\nMots : %d\nLongueur moyenne : %.2f\n",
		a.Words, a.AvgLen))

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, []byte(b.String()), 0o644)
}

func statsWords(lines []string) (int, float64) {
	var count, sum int
	for _, l := range lines {
		for _, tok := range strings.Fields(l) {
			if tok != "" && (tok[0] < '0' || tok[0] > '9') {
				count++
				sum += len(tok)
			}
		}
	}
	if count == 0 {
		return 0, 0
	}
	return count, float64(sum) / float64(count)
}

// FetchMany télécharge plusieurs articles en parallèle.
func FetchMany(titles []string) ([]*Article, error) {
	var (
		wg   sync.WaitGroup
		resC = make(chan *Article, len(titles))
		errC = make(chan error, len(titles))
	)

	for _, t := range titles {
		title := strings.TrimSpace(t)
		if title == "" {
			continue
		}
		wg.Add(1)
		go func(tt string) {
			defer wg.Done()
			a, err := Fetch(tt)
			if err != nil {
				errC <- err
				return
			}
			resC <- a
		}(title)
	}
	wg.Wait()
	close(resC)
	close(errC)

	if len(errC) > 0 {
		return nil, <-errC
	}
	var arts []*Article
	for a := range resC {
		arts = append(arts, a)
	}
	return arts, nil
}
