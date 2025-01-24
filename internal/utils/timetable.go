package utils

import (
	"fmt"
	"time"
)

type TimetableApiResponse struct {
	Status               string            `json:"status"`
	OperationMessageList interface{}       `json:"operationMessageList"`
	Result               []TimetableResult `json:"result"`
	TotalResultCount     int               `json:"totalResultCount"`
	Success              bool              `json:"success"`
	Fail                 bool              `json:"fail"`
}

type TimetableResult struct {
	SysNWAuthID               int    `json:"_SYS_NW_auth_id"`
	SysNWDataOd               string `json:"_SYS_NW_data_od"`
	SysNWDataDo               string `json:"_SYS_NW_data_do"`
	SysNWRokAkademickiRandki  string `json:"_SYS_NW_rok_akademicki_randki"`
	SysNWRokAkademickiBiezacy string `json:"_SYS_NW_rok_akademicki_biezacy"`
	Przedmiot                 string `json:"przedmiot"`
	TypPrzedmiotu             string `json:"typPrzedmiotu"`
	LiczbaGodzin              string `json:"liczbaGodzin"`
	DataZajec                 string `json:"dataZajec"`
	GodzinaOd                 string `json:"godzinaOd"`
	GodzinaDo                 string `json:"godzinaDo"`
	Dydaktyk                  string `json:"dydaktyk"`
	NazwaSali                 string `json:"nazwaSali"`
	Lokalizacja               string `json:"lokalizacja"`
	PoziomStudiowSkrot        string `json:"poziomStudiowSkrot"`
	Kierunek                  string `json:"kierunek"`
	Forma                     string `json:"forma"`
	Uwagi                     string `json:"uwagi"`
	Temat                     string `json:"temat"`
	Specjalnosc               string `json:"specjalnosc"`
	Grupa                     string `json:"grupa"`
	Wydzial                   string `json:"wydzial"`
}

type ParsedTimeTable struct {
	Status           string                               `json:"status"`
	TotalResultCount int                                  `json:"totalResultCount"`
	Success          bool                                 `json:"success"`
	Fail             bool                                 `json:"fail"`
	TimeTableEntries map[time.Time][]ParsedTimetableEntry `json:"timeTableEntries"`
}

type ParsedTimetableEntry struct {
	Lecturer       string    `json:"lecturer"`       // maps to Dydaktyk
	Subject        string    `json:"subject"`        // maps to Przedmiot
	CourseType     string    `json:"courseType"`     // maps to TypPrzedmiotu
	HoursCount     string    `json:"hoursCount"`     // maps to LiczbaGodzin
	Date           time.Time `json:"date"`           // maps to DataZajec
	StartTime      time.Time `json:"startTime"`      // maps to GodzinaOd
	EndTime        time.Time `json:"endTime"`        // maps to GodzinaDo
	RoomName       string    `json:"roomName"`       // maps to NazwaSali
	Location       string    `json:"location"`       // maps to Lokalizacja
	FieldOfStudy   string    `json:"fieldOfStudy"`   // maps to Kierunek
	Mode           string    `json:"mode"`           // maps to Forma
	Notes          string    `json:"notes"`          // maps to Uwagi
	Topic          string    `json:"topic"`          // maps to Temat
	Specialization string    `json:"specialization"` // maps to Specjalnosc
	Group          string    `json:"group"`          // maps to Grupa
	Faculty        string    `json:"faculty"`        // maps to Wydzial
	DayOfWeek      string    `json:"dayOfWeek"`      // adds what day of week it is
}

func ParseTimeTableData(tt *TimetableApiResponse) (*ParsedTimeTable, error) {
	var parsedTt ParsedTimeTable
	parsedEntries := make(map[time.Time][]ParsedTimetableEntry)

	for _, unParsedEntry := range tt.Result {
		parsedDate, err := time.Parse(time.DateOnly, unParsedEntry.DataZajec)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for entry: %v, error: %v", unParsedEntry.DataZajec, err)
		}
		dayOfWeek := parsedDate.Weekday().String()

		startTime, err := time.Parse("15:04", unParsedEntry.GodzinaOd)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for entry: %v, error: %v", unParsedEntry.DataZajec, err)
		}
		endTime, err := time.Parse("15:04", unParsedEntry.GodzinaDo)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for entry: %v, error: %v", unParsedEntry.DataZajec, err)
		}

		parsedEntry := ParsedTimetableEntry{
			Lecturer:       unParsedEntry.Dydaktyk,
			Subject:        unParsedEntry.Przedmiot,
			CourseType:     unParsedEntry.TypPrzedmiotu,
			HoursCount:     unParsedEntry.LiczbaGodzin,
			Date:           parsedDate,
			StartTime:      startTime,
			EndTime:        endTime,
			DayOfWeek:      dayOfWeek,
			RoomName:       unParsedEntry.NazwaSali,
			Location:       unParsedEntry.Lokalizacja,
			FieldOfStudy:   unParsedEntry.Kierunek,
			Mode:           unParsedEntry.Forma,
			Notes:          unParsedEntry.Uwagi,
			Topic:          unParsedEntry.Temat,
			Specialization: unParsedEntry.Specjalnosc,
			Group:          unParsedEntry.Grupa,
			Faculty:        unParsedEntry.Wydzial,
		}
		parsedEntries[parsedDate] = append(parsedEntries[parsedDate], parsedEntry)
	}

	parsedTt.Status = tt.Status
	parsedTt.TotalResultCount = tt.TotalResultCount
	parsedTt.Success = tt.Success
	parsedTt.Fail = tt.Fail
	parsedTt.TimeTableEntries = parsedEntries

	return &parsedTt, nil
}
