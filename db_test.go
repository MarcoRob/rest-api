package user_report

import (
	"testing"
	"log"
)

const (
	TODAY_TASKS = 0
	DELAYED_TASKS = -1
)

func testDB(t *testing.T, db UserReportDatabase) {
	defer db.Close()

	habitsArray, err := FetchUserHabits()
	if err != nil {
		log.Panic(err)
	}

	tasksArray, err := FetchUserTasks()
	if err != nil {
		log.Panic(err)
	}

	delayedTasks := extractTaskType(tasksArray, DELAYED_TASKS)
	todayTasks := extractTaskType(tasksArray, TODAY_TASKS)

	goodHabits := extractHabitType(habitsArray, "good")
	badHabits := extractHabitType(habitsArray, "bad")

	report := &UserReport{
		UserID: 201,
		TasksToday:todayTasks,
		TasksDelayed:delayedTasks,
		HabitsGood:goodHabits,
		HabitsBad:badHabits,
	}

	reportId, err := db.AddUserReport(report)
	if err != nil {
		t.Fatal(err)
	}

	report.ReportID = reportId
	gotReport, err := db.GetUserReport(reportId)
	if err != nil {
		t.Error(err)
	}
	if got, want := gotReport.UserID, report.UserID; got != want {
		t.Errorf("Update id: got %d, want %d", got, want)
	}

}

func TestMysqlDB(t *testing.T) {
	t.Parallel()
	db, err := newMySQLDB(MySQLConfig{
		Username: "root",
		Password: "admin",
		Host: "localhost",
		Port: 3306,
	})
	if err != nil {
		t.Fatal(err)
	}
	testDB(t, db)
}