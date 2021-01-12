package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"storj.io/common/uuid"
)

type reading struct {
	book     string
	passages []string
}

type bookBoundary struct {
	start, end int
}

var books = map[string]struct{}{
	"Genesis":              struct{}{},
	"Exodus":               struct{}{},
	"Leviticus":            struct{}{},
	"Numbers":              struct{}{},
	"Deuteronomy":          struct{}{},
	"Joshua":               struct{}{},
	"Judges":               struct{}{},
	"Ruth":                 struct{}{},
	"1 Samuel":             struct{}{},
	"2 Samuel":             struct{}{},
	"1 Kings":              struct{}{},
	"2 Kings":              struct{}{},
	"1 Chronicles":         struct{}{},
	"2 Chronicles":         struct{}{},
	"Ezra":                 struct{}{},
	"Nehemiah":             struct{}{},
	"Tobit":                struct{}{},
	"Judith":               struct{}{},
	"Esther":               struct{}{},
	"1 Maccabees":          struct{}{},
	"2 Maccabees":          struct{}{},
	"Job":                  struct{}{},
	"Psalms":               struct{}{},
	"Psalm":                struct{}{},
	"Proverbs":             struct{}{},
	"Ecclesiastes":         struct{}{},
	"Song of Songs":        struct{}{},
	"Song of Solomon":      struct{}{},
	"Wisdom":               struct{}{},
	"Sirach":               struct{}{},
	"Isaiah":               struct{}{},
	"Jeremiah":             struct{}{},
	"Lamentations":         struct{}{},
	"Baruch":               struct{}{},
	"Ezekiel":              struct{}{},
	"Daniel":               struct{}{},
	"Hosea":                struct{}{},
	"Joel":                 struct{}{},
	"Amos":                 struct{}{},
	"Obadiah":              struct{}{},
	"Jonah":                struct{}{},
	"Micah":                struct{}{},
	"Nahum":                struct{}{},
	"Habakkuk":             struct{}{},
	"Zephaniah":            struct{}{},
	"Haggai":               struct{}{},
	"Zechariah":            struct{}{},
	"Malachi":              struct{}{},
	"Matthew":              struct{}{},
	"Mark":                 struct{}{},
	"Luke":                 struct{}{},
	"John":                 struct{}{},
	"Acts":                 struct{}{},
	"Acts of the Apostles": struct{}{},
	"Romans":               struct{}{},
	"1 Corinthians":        struct{}{},
	"2 Corinthians":        struct{}{},
	"Galatians":            struct{}{},
	"Ephesians":            struct{}{},
	"Philippians":          struct{}{},
	"Colossians":           struct{}{},
	"1 Thessalonians":      struct{}{},
	"2 Thessalonians":      struct{}{},
	"1 Timothy":            struct{}{},
	"2 Timothy":            struct{}{},
	"Titus":                struct{}{},
	"Philemon":             struct{}{},
	"Hebrews":              struct{}{},
	"James":                struct{}{},
	"1 Peter":              struct{}{},
	"2 Peter":              struct{}{},
	"1 John":               struct{}{},
	"2 John":               struct{}{},
	"3 John":               struct{}{},
	"Jude":                 struct{}{},
	"Revelation":           struct{}{},
}

const header = `BEGIN:VCALENDAR
PRODID:-//Google Inc//Google Calendar 70.9054//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:Bible in a Year`

const footer = "END:VCALENDAR"

const event = `BEGIN:VEVENT
DTSTART;VALUE=DATE:%s
DTEND;VALUE=DATE:%s
RRULE:FREQ=YEARLY
DTSTAMP:20210112T151454Z
UID:%s
DESCRIPTION:%s
STATUS:CONFIRMED
SUMMARY:%s
TRANSP:OPAQUE
END:VEVENT`

// TODO(isaac):
// - Transp: TRANSPARENT

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 3 {
		return errors.New("Please provide path to plan file")
	}
	planpath := os.Args[1]
	icalpath := os.Args[2]

	planfile, err := os.Open(planpath)
	if err != nil {
		return err
	}
	defer planfile.Close()

	icalfile, err := os.Create(icalpath)
	if err != nil {
		return err
	}
	defer icalfile.Close()

	startDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	w := bufio.NewWriter(icalfile)
	defer func() { err = w.Flush() }()

	err = writeLine(w, header)
	if err != nil {
		return err
	}

	// period is the narrative period we're currently in
	var period string
	scanner := bufio.NewScanner(planfile)
	for scanner.Scan() {
		text := scanner.Text()

		if !strings.HasPrefix(text, "Day") {
			period = text
			continue
		}

		splits := strings.Split(text, " ")
		if len(splits) < 4 {
			return fmt.Errorf("Invalid line: expected at least 4 splits. Got: %s", text)
		}

		readings := convertToReadings(splits[2:])

		day, err := strconv.Atoi(splits[1])
		if err != nil {
			return err
		}

		dtstart := formatDate(startDate.AddDate(0, 0, day-1))
		dtend := formatDate(startDate.AddDate(0, 0, day))
		uid, err := uuid.New()
		if err != nil {
			return err
		}
		description := generateDescription(readings)
		summary := fmt.Sprintf("Day %s: %s", splits[1], period)
		err = writeLine(w, fmt.Sprintf(event, dtstart, dtend, uid.String(), description, summary))
		if err != nil {
			return err
		}
	}

	err = writeLine(w, footer)
	if err != nil {
		return err
	}

	return nil
}

func writeLine(w io.Writer, line string) error {
	_, err := w.Write([]byte(line))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	return err
}

func formatDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%s%s", year, fmt.Sprintf("%02d", month), fmt.Sprintf("%02d", day))
}

func generateDescription(readings []*reading) string {
	var s strings.Builder
	for i, reading := range readings {
		if i > 0 {
			s.WriteRune('\n')
			s.WriteString(" <br><br>")
		}
		s.WriteString(reading.book)
		s.WriteString(" ")
		s.WriteString(strings.Join(reading.passages, ", "))
	}
	s.WriteRune('\n')
	s.WriteString(" <br><br>")
	s.WriteString(fmt.Sprintf(`<a href="%s">Bible Gateway</a>`, generateBibleGatewayLink(readings)))
	return s.String()
}

func generateBibleGatewayLink(readings []*reading) string {
	bglink := `https://
 www.biblegateway.com/passage/?search=
 %s
 &version=ESV`
	var s strings.Builder
	for i, reading := range readings {
		if i > 0 {
			s.WriteString(url.PathEscape(";"))
		}
		s.WriteString(url.PathEscape(reading.book))
		s.WriteString("+")
		s.WriteString(url.PathEscape(strings.Join(reading.passages, ", ")))
	}
	return fmt.Sprintf(bglink, s.String())
}

func convertToReadings(raw []string) (readings []*reading) {
	fmt.Println("convertToReadings", raw)
	r := &reading{}
	for i := 0; i < len(raw); {
		token := raw[i]
		fmt.Println("token", token)
		var foundBook bool
		for j := 1; j < len(raw)-i; j++ {
			book := strings.Join(raw[i:i+j], " ")
			fmt.Println("check book", book)
			if _, foundBook = books[book]; foundBook {
				fmt.Println("found book!", book)
				r = &reading{book: book}
				readings = append(readings, r)
				i += j // increment correct amount here
				break
			}
		}
		if !foundBook {
			fmt.Println("didn't find book, appending token", token)
			r.passages = append(r.passages, token)
			i++
		}
	}
	fmt.Println("readings!", readings)
	return readings
}
