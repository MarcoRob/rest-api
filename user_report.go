package user_report

import (
	"net/http"
	"errors"
	"encoding/json"
	"time"
)

const (
	HABITS_URL = "https://habits-microservice-marcorob.c9users.io"
	TASKS_URL = "http://10.43.88.167:8080"
	TODAY_TASKS = 0
	DELAYED_TASKS = -1
)

type Habit struct {
	Difficulty 	string 	`json:"difficulty"`
	Color		string	`json:"color"`
	Score 		int 	`json:"score"`
	HabitID		string	`json:"_id"`
	UserID		string	`json:"userID"`
	Type 		string 	`json:"type"`
	Title		string	`json:"title"`
}

type Task struct {
	CompletedDate 	*int64		`json:"completedDate"`
	Description 	string 		`json:"description"`
	DueDate 		int64 		`json:"dueDate"`
	Reminder 		int64		`json:"remind"`
	Title 			string 		`json:"title"`
	UserID			string		`json:"userId"`
}

type UserReport struct {
	ReportID		int64	`json:"reportID"`
	UserID			string	`json:"userID"`
	TodayTasks 		[]Task 	`json:"todayTasks"`
	DelayedTasks 	[]Task 	`json:"delayedTasks"`
	GoodHabits		[]Habit `json:"goodHabits"`
	BadHabits		[]Habit `json:"badHabits"`
}

type UserReportDatabase interface {
	AddUserReport(*UserReport) (reportId int64, err error)

	GetUserReport(reportId int64) (*UserReport, error)

	Close()
}

func fetchUserHabits(userId string) ([]Habit, error) {
	//resp, err := http.Get(HABITS_URL)
	resp, err := http.Get(HABITS_URL + "/users/" + userId + "/habits")
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

func fetchUserTasks(userId string) ([]Task, error) {
	//resp, err := http.Get(TASKS_URL)
	resp, err := http.Get(TASKS_URL + "/Task/users/" + userId + "/tasks")
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

func extractHabitType(userId string, habitType string) ([]Habit, error) {
	allUserHabits, err := fetchUserHabits(userId)
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

func extractTaskType(userId string, taskType int) ([]Task, error) {
	allUserTasks, err := fetchUserTasks(userId)
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

func GenerateUserReport(userId string) (UserReport, error) {
	var err error
	report := UserReport{UserID:userId}

	report.DelayedTasks, err = extractTaskType(userId, DELAYED_TASKS)
	if err != nil {
		return UserReport{}, err
	}
	report.TodayTasks, err = extractTaskType(userId, TODAY_TASKS)
	if err != nil {
		return UserReport{}, err
	}
	report.GoodHabits, err = extractHabitType(userId, "good")
	if err != nil {
		return UserReport{}, err
	}
	report.BadHabits, err = extractHabitType(userId,"bad")
	if err != nil {
		return UserReport{}, err
	}

	return report, nil
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