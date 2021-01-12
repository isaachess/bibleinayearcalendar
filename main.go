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
TRANSP:TRANSPARENT
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
		text := strings.ReplaceAll(scanner.Text(), ",", "")

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
		uid, ok := uids[day]
		if !ok {
			return fmt.Errorf("UID not found for %d", day)
		}
		if err != nil {
			return err
		}
		description := generateDescription(readings)
		summary := fmt.Sprintf("Day %s: %s", splits[1], period)
		err = writeLine(w, fmt.Sprintf(event, dtstart, dtend, uid, description, summary))
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
	r := &reading{}
	for i := 0; i < len(raw); {
		token := raw[i]
		var foundBook bool
		for j := 1; j < len(raw)-i; j++ {
			book := strings.Join(raw[i:i+j], " ")
			if _, foundBook = books[book]; foundBook {
				r = &reading{book: book}
				readings = append(readings, r)
				i += j // increment correct amount here
				break
			}
		}
		if !foundBook {
			r.passages = append(r.passages, token)
			i++
		}
	}
	return readings
}

var uids = map[int]string{
	1:   "f01d949f-576d-4f0b-bfb8-24204bacdd77",
	2:   "dbeb76c5-1651-459d-b6ec-17be645f9865",
	3:   "f50687df-74a4-4537-b3fd-d2d0acc67ce3",
	4:   "7d590c7a-33ef-49bf-bdd8-24ce6518dc3a",
	5:   "205cfcc9-e3df-43c2-8af3-18289e4f216c",
	6:   "160786b2-05a5-4c9e-9659-1c9addc12eea",
	7:   "1f66c473-f736-4c9e-9d03-bf398b3ce6a9",
	8:   "6a7cd07d-4ad3-4522-a17b-ebb7d9c8a65c",
	9:   "6850a53e-6848-4e42-afea-4db20333f503",
	10:  "9e705da8-1a54-496e-be64-b8dba944a541",
	11:  "5a71eb99-8932-47e5-983d-0c8be08c34a0",
	12:  "4290a7cc-b6da-49e8-9a4e-d0fabb8f1279",
	13:  "d4d9ee31-6b84-459d-80f3-b0ed1b3b34b9",
	14:  "c1b33363-3252-41b3-afee-4b53fae5c362",
	15:  "40c68cd1-a49c-4eb9-bb55-89df6bd23401",
	16:  "e040590b-8611-4e21-9aff-9f7dac6e4964",
	17:  "d7276a37-ddb1-4e44-95d7-26216ef716b1",
	18:  "4c50b23e-1e47-4822-b0b7-326671ed8459",
	19:  "c5cd0bef-7667-4afe-812b-31ae36535468",
	20:  "244c78d9-0ec2-48d2-bb29-874b48b1678f",
	21:  "61a9c517-2a05-4fc7-9945-536da2849de3",
	22:  "463c84e2-77ac-4b7a-acf4-c4c0785c8f4e",
	23:  "68085dd4-9323-483d-94d6-df49e7441852",
	24:  "e7627980-a290-4c7e-9087-903c02f4d5b3",
	25:  "bc71f523-a5df-492c-8c3c-a0c82a9939b6",
	26:  "20233479-f446-46a6-af8e-c5c68b4ab5fc",
	27:  "8848d29b-1fed-4127-9ec9-906dc5c0b34a",
	28:  "9dd05454-75ec-4f65-a99c-7c2c1f0329e8",
	29:  "47e46a5e-a947-4310-a010-e947a9c4103e",
	30:  "1352a717-ed5b-4477-b494-ec3fd8267af9",
	31:  "79172194-2ed2-4e57-b4b4-3080bfccc0cc",
	32:  "8750aac3-6e5b-4c4f-a354-13a174f06df9",
	33:  "29e572b5-3f5d-4fc8-9b27-dececf290741",
	34:  "0b084aa8-850b-4da7-be85-a803ec1b9d13",
	35:  "73beb186-362f-47c4-a287-9164c9b9072f",
	36:  "2ac73554-e929-4478-99b4-04c61071648e",
	37:  "57fb2a56-21b5-4a72-b3bb-7e219f9f852a",
	38:  "f0fc5d86-641a-4779-9251-3d4ef2869c58",
	39:  "751f7af9-599d-4006-941f-543950d39645",
	40:  "4b1a6ff8-2bdd-4325-8490-1f62218b23d1",
	41:  "c11c79ff-7867-4f36-b047-c13f058537bb",
	42:  "8032efc1-7bf7-4335-8c8e-3c1b926920f0",
	43:  "093c1f16-3e06-4afb-a354-c588b5d82cdb",
	44:  "04ce6abd-9c16-46bf-ba23-2acf6f23dc40",
	45:  "3cd47581-84d2-4497-883e-fd8451d77827",
	46:  "288b9a19-fbad-4aa4-8f32-b5298f1a8206",
	47:  "2ae0983b-2d1e-473d-816b-ecb110a6e2f0",
	48:  "ef31305a-4a7d-417e-94f8-7a9e5ca3422a",
	49:  "0d8c14b7-a3a5-4fa7-a152-70746be98e17",
	50:  "b06cb52c-fdcb-4539-b310-d3bb86596ab6",
	51:  "faf2da7e-9b00-427d-85db-54af5fd2b3cc",
	52:  "e4c6cd3c-bfe2-4f64-848b-9b48b9ddd8d3",
	53:  "ea30b711-10e3-471d-ba3e-d7325e2e4d6e",
	54:  "5207d0a2-900b-4473-a2a2-0b0c4e8c212b",
	55:  "0b845cfa-6ee6-4b34-9f9f-46b6a98d3a19",
	56:  "b36cf42c-b49f-4f8e-af9f-f7b9142ee903",
	57:  "a3a6762f-9e57-4939-801c-d4bacd1252b9",
	58:  "0946427e-6bfa-4e86-be8b-8950e7f6b413",
	59:  "a7bbc386-f28a-4899-a7c8-851429964094",
	60:  "f29bdeeb-75d4-42a8-9c32-ad2d96c3140a",
	61:  "995131b8-7b8a-4327-a772-c769deb50a50",
	62:  "1f7ba216-9bce-44fe-92f8-d7f2de527b42",
	63:  "f5e97370-1e6b-4767-a067-93933a21a027",
	64:  "b2f5eb5d-74fc-438f-a9c5-557c3960b6d9",
	65:  "2db0e92c-f680-447c-868e-19fcd19415fb",
	66:  "4e78eaa0-4daf-4f26-9080-f942f34d7f6c",
	67:  "49bb300f-5bb3-4199-959e-a56af659d184",
	68:  "ebf1e133-aa9f-4f5d-be42-8f75a0d7cbe4",
	69:  "b2c75de1-c1d5-480e-aa22-b22a9a130eae",
	70:  "f4e5ac78-b386-462e-ad2a-b0fa8b7af166",
	71:  "df93b33c-ed0c-4a67-8494-84e373712738",
	72:  "c790c9ef-1434-41e1-948a-0eca18785cc6",
	73:  "8c4b8d65-387a-4968-bb64-b8c8e5c0680b",
	74:  "c2725f80-ada4-413d-a2ac-ff4f311300bd",
	75:  "922939f7-5260-455b-a3c8-7c9b006ba868",
	76:  "1bd6d567-b523-4767-b57b-eb19ed7013ff",
	77:  "198cf2d6-915e-4033-bdda-2f44cc2a8101",
	78:  "3df49cca-9c99-43e1-93ea-c5119c595917",
	79:  "d834babf-b537-41a9-abfd-0fd267d245e7",
	80:  "f2fc831d-0186-45c1-960f-84b906c7a30a",
	81:  "3cbd7679-d993-4d86-8f35-11eeeee64bb2",
	82:  "0f2beb71-e89d-4ff5-8c1b-a41336cfb6df",
	83:  "61244b71-8039-4cb5-a51a-870f14cc8b01",
	84:  "2a147c53-7635-4a68-8664-ed13a98ffc04",
	85:  "d93f82c1-8964-4856-9ed0-ae71805e9126",
	86:  "70ac347d-3e6c-4b24-91c2-e21fc6457518",
	87:  "4fef788c-7816-4ab4-bc32-1ad83af6991e",
	88:  "009b0275-0fab-4f12-bd90-11acffbd8416",
	89:  "36f121e9-9d5f-4a67-9a41-03da51aef8aa",
	90:  "2ed8622c-ea2f-4144-a22d-04ea3e1c4503",
	91:  "f437b1d9-2a93-470a-b6e0-3e129aaa4c2e",
	92:  "96cf1ee1-bd1f-48e5-af6f-f3d7dd2aafa6",
	93:  "f02aba0b-9964-477f-9302-3937b1c0baff",
	94:  "1f3c7bd6-3078-49aa-a7cc-b389205d4fb7",
	95:  "eae00b3c-3ac5-44bd-b22c-22c772d0a115",
	96:  "30e5b8a7-7679-4518-b91f-4b2eed80ed28",
	97:  "0dbd5b61-717f-4a0f-af36-cb053f79345c",
	98:  "a9cce97b-38fc-4aa9-a038-4eff93540714",
	99:  "6a75c3b1-6b59-46e9-9f08-3c9ae215239d",
	100: "1cf1b795-0ecd-4869-9032-f61116b912d2",
	101: "ecb537b3-282f-4e92-a90f-77149322ca4f",
	102: "c085e385-f666-4343-a90a-5684a30eb130",
	103: "f479b6cb-93f4-4d31-a749-1e94748fb64e",
	104: "ddf29ad6-d864-4dc8-addf-3ca9d58ac6b0",
	105: "33a5db72-1249-4af3-a5c5-5dccc56f209b",
	106: "746f5b6c-55a3-451f-89d3-d949439b8ece",
	107: "7c161408-9081-4625-825e-26c00bec4157",
	108: "9b680912-eacf-4820-8835-ddf2808ad341",
	109: "9cfee9a2-dc4a-4c4e-a75d-c36a32ccc16b",
	110: "3f8a1b6c-ad4b-4158-adda-5b1f24e41ad4",
	111: "8b3f6d50-324e-4b24-8e14-1a5f71388335",
	112: "2625dbdc-0a55-4a3a-906d-d6d088f0277d",
	113: "00aedb79-6564-4778-b6a0-e27f8370975a",
	114: "686fb772-7896-481f-8b8d-68e8c2e90cc6",
	115: "372193ae-31fc-4e9e-8264-3195abaac3db",
	116: "c363fa52-2426-493c-9c84-f23192c2b230",
	117: "a078a90d-69d2-48d9-a6a4-5b887b781e9f",
	118: "009a3761-60ba-456e-b373-00df18cf684c",
	119: "e81dc23c-037c-4803-8910-5190a58e3b33",
	120: "29267b27-47a9-467d-aaa6-ff79e865b332",
	121: "670be518-810f-4287-a9c5-b75bf57b4bd8",
	122: "88434570-8b80-4f46-b964-cfbdb7ff69cb",
	123: "391b16f9-0f24-48bb-94b1-9b218e0b37e3",
	124: "4ae4ba71-f12b-4311-ac09-3ce0ac7bb3c3",
	125: "3ee5d638-2c82-4975-a4dc-80f35188b189",
	126: "4ab07148-aaa7-47c5-b657-6536f2dc55dd",
	127: "8fc27369-6687-4540-8d85-5c78f272ce5a",
	128: "9a9f6094-d03c-4886-baab-88e2aa16fea7",
	129: "a46a0294-0d12-46c5-a1c6-9c63c1345f02",
	130: "067dfe40-49f9-4cb5-9841-4d049452acee",
	131: "62034783-badb-4db4-8185-64c3745db02d",
	132: "910bfb8c-1aa8-43b8-b631-70a41abc06dc",
	133: "bec22bd5-50f6-4511-bb25-d5f63c9b07fb",
	134: "cc7ad40c-736e-41c8-a2fa-80893fa7d9d4",
	135: "ba4a548b-786d-4c06-811f-4e0a71862d2e",
	136: "2eb5d7b3-76ae-447e-8b67-b15008fe3e5e",
	137: "035967e9-38d1-4215-a09d-de00af6063f0",
	138: "ae112472-550a-426c-80d0-4f0681f8acdd",
	139: "bad482c8-69e2-48f1-808d-1520a67c3905",
	140: "c1e451af-f401-47d9-8273-663f6a6efba2",
	141: "afe7c58a-a819-44f6-9cdb-0e533e5c2771",
	142: "36647c59-ef6e-44c0-85e9-01f120310240",
	143: "9dee1f03-b291-4a7f-b879-35be56414010",
	144: "6480f916-df3a-4a7f-9804-3e4201b23d81",
	145: "94c18f64-ed1d-497b-ad60-5ab2f04174f0",
	146: "d05faa1c-1c52-450a-938f-e932c5dc8e0b",
	147: "e37945be-5209-4c1b-b703-119bcd23df03",
	148: "07e30351-0581-4d3e-9069-b4cb72c4de9e",
	149: "e97b252a-bb74-4f34-b9e0-7b9b3d4d7b46",
	150: "cefb13e1-e385-4687-b6d0-c05923c0004e",
	151: "56f72921-8a96-4987-9d03-bb250fbd533f",
	152: "827181f5-9bb8-46d9-9470-75f01c615679",
	153: "79333036-b3b2-48b3-8a86-a4d7f56b67a4",
	154: "fe85ea7e-dd0e-4d5a-9915-50ee30421cad",
	155: "755390d1-bb5e-444e-ada2-08e469cb1866",
	156: "a82e56d9-bb6f-4958-830a-c0c512cd102d",
	157: "ed7632c1-2b22-4e1f-aa5a-00e173310d05",
	158: "25f1a4a2-8d3c-43e2-b6ef-63d0076bb755",
	159: "b5f73391-7748-4aec-9af6-0622c375b0a1",
	160: "078ec58d-3f57-4666-b2c4-773fe95b450e",
	161: "62c60516-3530-4103-924f-14428d39510c",
	162: "c5350005-ef4a-4487-b36d-7b077bd49fb7",
	163: "a9ffbd6d-9e07-4062-b083-2b220d784417",
	164: "671a504e-64f5-4a53-9750-0759544e2eda",
	165: "f79ff309-34dc-4778-807a-3a8c0671dd4c",
	166: "e23a4f2c-3993-4e59-bfe9-b88b2a5f4ff9",
	167: "0b656025-d3a4-42e8-9bd6-d3f95b3123d6",
	168: "2b9fdd2b-5f59-42c5-bc11-b6e5d05569ab",
	169: "9d928955-b2df-48ca-a14f-34d1c4b37b1b",
	170: "8c266db9-3e64-4cf4-aa23-a41d724a33aa",
	171: "59b28353-d683-4aa5-815d-76c694483002",
	172: "34c9f0ca-dbbb-4f8e-ac90-216691f4ded2",
	173: "94c2e23b-5411-41ed-ab9a-9b0f633a0ee0",
	174: "82ebcb30-dce9-4430-98df-666a2771d1c9",
	175: "d7d2113e-05d3-40d6-b7c5-c0beebcde807",
	176: "d8bbb6eb-c00e-48b4-be4f-81c9d291363a",
	177: "407b394e-5548-4104-9fde-71e0fe366032",
	178: "d3ecfbe6-5d6e-4179-8add-f9a6b3dd4312",
	179: "65755914-f061-4085-8dcc-8baaaa28e5ad",
	180: "c2865d5e-4757-4c1d-b949-37fcb9d1042e",
	181: "9ce52b10-d5d5-4f26-8fc9-88a0f35b5135",
	182: "6645f506-51dc-4d45-9528-c3c3b394cbf6",
	183: "e02dfdf7-bc05-4f22-b5a6-9f2cd5782f9f",
	184: "eda92185-0c48-4fac-9c05-b6b8e038da8f",
	185: "8ac3c1c6-7975-4094-ae7b-83970e011b13",
	186: "b8735a34-adbd-447a-aaf0-09adf80e786a",
	187: "eb740afb-63d6-46a8-aba5-77804fcd6030",
	188: "5985b157-86a0-4a5b-923f-c7c608bbddc6",
	189: "937b64ad-8460-462f-a97c-e4af782a2b08",
	190: "b240a39a-3d3d-4ed8-9fa8-4f6d3b8501e2",
	191: "f5531c71-73b6-4e6c-8e77-9e7fca18e06c",
	192: "7adcab0c-7d93-41ef-8c07-5310b6378be4",
	193: "d3d7b5ea-b566-4c9e-a29a-37df6bfe8ac4",
	194: "e94e0208-3e03-41c8-87e0-55eb14a9e3fe",
	195: "b9378946-b9c9-4a0b-bec6-c3d38e4d8a00",
	196: "2e128e39-3e86-4aac-aadb-27f2ffaabfed",
	197: "8c2805e1-3d61-4d84-a993-87f81a9cd63f",
	198: "23684388-ae8a-453f-a2cb-0513162c0074",
	199: "1bee41c9-7f40-446d-bfa3-162dead98873",
	200: "d1f7283b-6f2b-4070-b2fe-a63366116185",
	201: "e5b90d7a-80ee-4169-b23a-478d6465138e",
	202: "0a274d7c-146f-4ffd-94df-83660aa17c02",
	203: "ed76fd61-1107-48a0-94fe-b8c35555b356",
	204: "2151b694-2a0f-4d03-b046-090b0316342c",
	205: "3cf9266e-2f88-4f34-be8a-0f24f72bd321",
	206: "4ac770ca-35e7-4a29-b393-874862236c76",
	207: "3284f17c-e1a3-42b8-a7c7-bde55cfd1067",
	208: "412c71c8-4d5f-4ee5-bc78-152176637a31",
	209: "2f041c12-787e-428a-a690-a00a66d2f32e",
	210: "1b97eaa4-8335-4590-8f5d-af4aa174b904",
	211: "c6edf805-098e-4c24-b8c3-886b10ce5cb5",
	212: "836806ec-7b38-4ec8-9ff6-7a532c80e3a4",
	213: "bf2def9a-4a5f-460f-bd2b-d99549ef8cd3",
	214: "aad22d69-1689-4bcd-8176-dfe9a8f3e9cb",
	215: "23857c15-239c-4c35-b19b-ba4c763863da",
	216: "9932b17f-90c2-41bc-9362-bcdea8c6b363",
	217: "cc9888f9-8aba-4f61-baff-a5ea9fd22ed5",
	218: "0e21951c-8b2d-4271-87e2-4ee686f56b9b",
	219: "9c735132-15f0-47f9-a487-17bee3191dfe",
	220: "5e341190-020b-4a4c-8903-0d7af1c02779",
	221: "8c388934-e395-4a91-9be7-b64b2e2075c9",
	222: "509396de-3506-4d3f-94f7-4656b140c590",
	223: "a2358cd6-0dab-46e3-9e1e-42738ffb881a",
	224: "0f7a972d-6724-4875-b56a-dfea3238d723",
	225: "310f34ec-b001-450c-ada3-483f100d7854",
	226: "a26c7607-ffdd-4832-bb13-d56dbbe9e4f2",
	227: "02b0eed3-80b9-4b17-a2d9-8dc61326d54a",
	228: "b6f1297c-6797-4dde-a50d-37abe7ec4989",
	229: "7a7fe111-0ab2-47ac-a943-9cd7250e5531",
	230: "6e60b45e-cbad-43f4-89e5-5b656da2b383",
	231: "31819806-6b7a-4d0e-a898-96e5efa4225a",
	232: "976c2d7c-7734-4e14-82dd-6ddd0e6b8329",
	233: "9d599158-06c9-44d6-9d31-54b0214a82c5",
	234: "d69ed7d9-93b7-455f-a582-82af86622ca5",
	235: "d7d72382-55d8-4280-b2a7-1dd836b4098b",
	236: "77d83ccf-1014-4bce-a464-99d30f310594",
	237: "a85d9f5a-1154-4394-b423-db784ff85c8b",
	238: "429af281-5393-4f16-ab27-5b818ae6897d",
	239: "5c1c5fed-3fe1-4bbb-83aa-80ec37fa82d1",
	240: "2052609a-3024-4630-a670-1a8d7e063eda",
	241: "478e0930-fad4-482f-a572-9f6447e5ca54",
	242: "4abc5764-24e1-4885-9675-6bdf39454ee5",
	243: "ffcf172a-efed-48e4-927d-2e89943aac1c",
	244: "c3ab2c0f-4cd6-4930-b38a-16d4c9caff82",
	245: "f6d1886b-87e1-4a49-8313-30eccaf97959",
	246: "60b65bbd-c303-409f-ae2f-d963caaafdd8",
	247: "7cf00fd9-d020-4940-8c03-36b4ab525855",
	248: "bf67c362-af4c-44f2-bc10-c98b51eab4a0",
	249: "9fa11fa2-5d75-4e15-8a61-610870031d7d",
	250: "39f31709-5c9b-4e9a-90b6-1dca0352d70c",
	251: "c357e41c-7b88-425c-b446-9a582c237b8f",
	252: "f388d98f-db55-4a5b-8023-f03ced889cd6",
	253: "1095d165-975b-4e53-bc31-dac9631e6993",
	254: "c860bd43-0ba2-40c2-9b2c-7f646da2d178",
	255: "b5aed7f8-783d-4dce-9c72-6557d71d5408",
	256: "a45064ba-4002-488e-abd2-ffd5f721cf21",
	257: "307438ed-565c-4933-9523-6b7f9a2bdc06",
	258: "5d62e468-1e06-4817-8226-d7d040e4fbb6",
	259: "224927bb-c288-454e-8203-9533819ba4f7",
	260: "b700d8a8-aaa6-486f-ba5b-add963ceb9ab",
	261: "cd1a2d2d-22bf-41b7-916b-8561b22d63db",
	262: "99511833-4c7d-4592-8023-12cf094cf670",
	263: "5ce764c0-01c7-4079-b273-58f939739b05",
	264: "ff4bc1e0-d132-404b-b149-839fe6c2a2c2",
	265: "89397b65-c0d1-4573-8cc1-a4a0d09b87fc",
	266: "4e0980ed-8f5e-49b1-91e9-e5f8857ae5e1",
	267: "b56fd4d0-be62-4167-afd0-8099a068b43f",
	268: "5194f2cf-502c-4947-925f-1d24528d8790",
	269: "358c805e-1619-4db4-acb7-686d0b693d09",
	270: "7639cf0f-9c1c-4700-b48b-03df37afc740",
	271: "eb685f0d-bbec-48f5-b4c4-1460eb465bca",
	272: "45d55ce8-ab9a-433a-a10c-b31c0716e47c",
	273: "8713b90d-d48b-4d52-af68-2ca805e89a77",
	274: "d72015f2-11a8-462a-83db-614cd32900b2",
	275: "f6853c55-0eea-4acf-83e9-39ed1e2d35f9",
	276: "1c5b6ece-ad3a-4e0c-8a61-781c4d250596",
	277: "86b4d655-3672-495c-8be4-bc8315790122",
	278: "361cf7ef-1b47-456b-b779-eeba36f3f5e0",
	279: "89ffc5b6-8dcf-4fa4-927b-39d250e0c0ab",
	280: "22efd437-2453-4efb-a7c5-eb5cea53703d",
	281: "a4c19d0b-6992-42f6-878c-9128370f50c1",
	282: "a07e80fb-c455-4dfd-bca0-14634db34bb7",
	283: "1addbf2e-5677-4290-90c7-4607f8f0ca6d",
	284: "6c09041b-0e55-4ff2-bf7b-b5e0775e2a83",
	285: "20b07a5d-4c63-42bb-8fd7-4fbb140465ae",
	286: "de701661-0b6b-440d-84e7-04f7ce230747",
	287: "6b4dccf3-cdf9-4324-bc6f-2ba47b7425cc",
	288: "820c934d-393e-4954-acd7-ec33434378de",
	289: "48b581b6-d101-4d6f-b45d-ef8ac4dab2c7",
	290: "16497e1d-baf6-4b64-a5e2-ea1bb049b6c1",
	291: "e33e8793-636f-4b97-915f-b666755d6540",
	292: "91631918-3746-463f-ab39-8becc73458c8",
	293: "db1f7914-df0b-4cb6-8bec-82cb1e6d41a6",
	294: "3d0e5bcd-d9ed-41de-b17d-4dff30b6bcf8",
	295: "93fc286a-fc7f-4bdc-82d2-cb41cddeb4b0",
	296: "7f635e79-38f0-4ee8-b11c-ef8fb46db0f1",
	297: "956d61ca-f61d-426d-889a-fb3a353dfc02",
	298: "f18272dd-42ec-407a-b75a-8adda79fb1c8",
	299: "7e22412a-72fe-4c28-95ec-725437f36ba0",
	300: "d3e22e68-b185-49e3-952e-423e74c71979",
	301: "5de99a3e-7428-4fb1-97e6-254e9820d84d",
	302: "8ed6a48a-2a6f-444c-8988-08e2c671ac14",
	303: "ff297881-eecb-429e-9686-ff2c680c63f4",
	304: "a75adb9c-7eb5-4f8c-882b-50a5e2a2519b",
	305: "b4837d32-84ce-4345-8ed1-5827ed52a91a",
	306: "478084c0-8d48-44c3-9ebc-7186ff63a5fb",
	307: "b71dadb5-b6f7-4892-b923-304e65afe652",
	308: "c3422f47-c864-454a-9197-ef5128602163",
	309: "75a8e777-8098-4fd2-9639-96ffc11581d4",
	310: "835c87b1-01e9-4f23-9973-9a4a38fb8eba",
	311: "9cf7b57d-5045-48c1-8804-1cee3b425669",
	312: "d075b7ec-4e4e-4cba-9ff7-1ddb0cc4174c",
	313: "75ca1114-bc73-46b0-817e-1dd76d05ba26",
	314: "b7ac9560-8fc9-4256-b3f5-21b3ad0ef4ae",
	315: "b29df1ae-d149-4b5b-9dbb-e7587d9605d8",
	316: "7d80dc7f-6fc5-403d-b5d1-343179edd627",
	317: "3b78dd8a-bea8-4a63-800f-7b03cf4583cd",
	318: "19de5ebb-053f-4e80-a78f-aef43de10b10",
	319: "36768796-a1c8-4840-9e26-2d9e389c2f2a",
	320: "75744dcb-24f5-4f41-a234-c51570f6755a",
	321: "59ef8895-dd3a-421b-81f9-0ad2850832a9",
	322: "d0bb885d-8b1a-448c-9f9a-76b607e5ca03",
	323: "c4c82539-a467-4c59-9a8e-081bb766084a",
	324: "d4bfd79f-2cc9-479f-a10b-df11500d1eb6",
	325: "9b26e7d9-9551-49c5-b74f-355896dceb18",
	326: "69584155-0254-4cfe-8ed0-c0f7164b127e",
	327: "e0f649cb-a5b0-4a08-bf9c-bf4091ca122a",
	328: "3244499b-3811-4256-afb6-f9e2112e4eaa",
	329: "8d6d8ef6-02d3-4323-afb0-612468bc1fcf",
	330: "bf920edb-1a59-4cbc-b0eb-bbbbed9feba4",
	331: "37e343d5-d8f2-4ba0-a948-e2259bd5754f",
	332: "0b2a60d4-d277-46bf-91e7-0dee80ece772",
	333: "1b460076-f7bc-4779-90d3-bde4c4e44d64",
	334: "c409c83e-8f22-4b02-9ec8-cfed5be4bf94",
	335: "16fd4dd5-f1be-47cf-8ad9-1e834b21c66f",
	336: "30dd0efb-4352-4866-8a18-59882eba546e",
	337: "c86e4a58-611b-46a1-82f9-3c45af2278cb",
	338: "964d8e38-b9a9-4e2d-b478-9ca90e13b746",
	339: "e12c4f13-66d8-4179-b55a-e9f163dc55cc",
	340: "6172e1dd-1eeb-472f-871d-4451540403db",
	341: "0938a521-5fd9-4201-963d-d4f7eccb9964",
	342: "0f7bb550-995b-49c5-8d8c-9f87ee6b0493",
	343: "f2a75903-d310-419b-8ec2-c090a18c089c",
	344: "bc51c016-bf9e-4bb5-a6f6-b6eb99c6ac23",
	345: "7a3fcb1d-02be-4304-bfd6-26dad3824b2f",
	346: "f6c4c341-2d04-4715-ba70-e4cfbb229f1e",
	347: "3b23d189-e497-4ed5-8cbc-3e9a22373853",
	348: "c29dce05-ea28-48c5-b1ea-106f8f76b503",
	349: "1b3ade4f-241e-4d12-8734-be29112acde1",
	350: "cf8c584a-aa69-4379-b085-976f86d40ae0",
	351: "d9deef10-48d1-469f-9aea-5e8fbe2d6081",
	352: "91832e2f-b080-46be-a45b-42325e6ebf4e",
	353: "d938c3d2-5f1f-4538-9fc8-72d417edf5d6",
	354: "e023f4f1-97ad-46af-9a51-b76ee0569576",
	355: "5e614faf-4911-4770-b1c0-9c21f7454ed0",
	356: "bc54963f-1a21-43d5-a7e1-376c7e71fcfc",
	357: "4f247793-5c85-4fe9-a34a-7e88b3c05795",
	358: "87d519c1-a0bf-499a-ad33-18a4ff360305",
	359: "8404f8be-c6d5-4ab9-8cdd-b9858cf066be",
	360: "9d6fab6e-5c40-43d2-b6ad-aa25747008de",
	361: "0a723251-4c0a-4516-a88a-e6c178fc69b1",
	362: "eedce24f-686d-4927-8ff5-ff2fd680d4f5",
	363: "5b99026f-1b5f-423d-8c99-af75ae45b2cf",
	364: "9a94227c-5531-41a5-8e42-8d12c5f8fef3",
	365: "59d3a354-71db-4ec4-8e97-cfc330494d46",
}
