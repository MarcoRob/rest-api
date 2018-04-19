package user_report

import (
	"net/http"
	"errors"
	"encoding/json"
	"time"
)

const HABITS_URL = "https://api.myjson.com/bins/12n767"
const TASKS_URL = "https://api.myjson.com/bins/bweof"

type Habit struct {
	ID			int64	`json:"_id"`
	Type 		string 	`json:"type"`
	Difficulty 	string 	`json:"difficulty"`
	Color		string	`json:"color"`
	Title		string	`json:"title"`
	Score 		int 	`json:"score"`
}

type Task struct {
	Title 		string 		`json:"title"`
	Description string 		`json:"description"`
	DueDate 	int64 		`json:"due_date"`
	Reminder 	Reminder 	`json:"reminder"`
}

type Reminder struct {
	Time string `json:"time"`
	Days int	`json:"days"`
}

type UserReport struct {
	ReportID		int64	`json:"report_id"`
	UserID			int64	`json:"_id"`
	TasksToday 		[]Task 	`json:"tasks_today"`
	TasksDelayed 	[]Task 	`json:"tasks_delayed"`
	HabitsGood 		[]Habit `json:"habits_good"`
	HabitsBad 		[]Habit `json:"habits_bad"`
}

type UserReportDatabase interface {
	AddUserReport(*UserReport) (reportId int64, err error)

	GetUserReport(reportId int64) (*UserReport, error)

	Close()
}

func FetchUserHabits() ([]Habit, error) {
	resp, err := http.Get(HABITS_URL)
	//resp, err := http.Get(HABITS_URL + "/habits/" + userId)
	if err != nil {
		return []Habit{}, errors.New("Habits unavailable")
	}

	jsonArray := make([]Habit, 0)
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err = decoder.Decode(&jsonArray); err != nil {
		return []Habit{}, errors.New("Error decoding Habits")
	}

	return jsonArray, nil
}

func FetchUserTasks() ([]Task, error) {
	resp, err := http.Get(TASKS_URL)
	//resp, err := http.Get(TASKS_URL + "/tasks/" + userId)
	if err != nil {
		return []Task{}, errors.New("Tasks unavailable")
	}

	jsonArray := make([]Task, 0)
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err = decoder.Decode(&jsonArray); err != nil {
		return []Task{}, errors.New("Error decoding Tasks")
	}

	return jsonArray, nil
}

func ExtractHabitType(habitType string) ([]Habit, error) {
	allUserHabits, err := FetchUserHabits()
	if err != nil {
		return []Habit{}, err
	}

	returnHabits := make([]Habit, 0)
	for i, _ := range allUserHabits {
		if allUserHabits[i].Type == habitType {
			returnHabits = append(returnHabits, allUserHabits[i])
		}
	}
	return returnHabits, nil
}

func ExtractTaskType(taskType int) ([]Task, error) {
	allUserTasks, err := FetchUserTasks()
	if err != nil {
		return []Task{}, err
	}

	returnTasks := make([]Task, 0)
	for i, _ := range allUserTasks {
		taskDate := time.Unix(allUserTasks[i].DueDate, 0)
		if compareDay(taskDate, time.Now()) == taskType {
			returnTasks = append(returnTasks, allUserTasks[i])
		}
	}
	return returnTasks, nil
}

// returns 0 if times on same day
// returns -1 if time1 is any day before time2
// returns 1 if time1 is any day after time2
func compareDay(time1 time.Time, time2 time.Time) (int) {
	//fmt.Printf("time1 day: %v, month: %v, year: %v\n", time1.Day(), time1.Month(), time1.Year())
	//fmt.Printf("time2 day: %v, month: %v, year: %v\n\n", time2.Day(), time2.Month(), time2.Year())
	if time1.Day() == time2.Day() && time1.Month() == time2.Month() && time1.Year() == time2.Year() {
		return 0
	}
	if time1.Unix() < time2.Unix() {
		return -1
	} else {
		return 1
	}
}