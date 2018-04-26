package user_report

import (
	"testing"
)

const (
	TODAY_TASKS = 0
	DELAYED_TASKS = -1
)

func testDB(t *testing.T, db UserReportDatabase) {
	defer db.Close()

	delayedTasks, err := ExtractTaskType(DELAYED_TASKS)
	if err != nil {
		t.Errorf("test failed with err: %v", err)
	}
	todayTasks, err := ExtractTaskType(TODAY_TASKS)
	if err != nil {
		t.Errorf("test failed with err: %v", err)
	}

	goodHabits, err := ExtractHabitType("good")
	if err != nil {
		t.Errorf("test failed with err: %v", err)
	}
	badHabits, err := ExtractHabitType("bad")
	if err != nil {
		t.Errorf("test failed with err: %v", err)
	}

	report := &UserReport{
		UserID: 201,
		TasksToday:todayTasks,
		TasksDelayed:delayedTasks,
		HabitsGood:goodHabits,
		HabitsBad:badHabits,
	}

	reportId, err := db.AddUserReport(report)
	if err != nil {
		t.Errorf("test failed with err: %v", err)
	}

	report.ReportID = reportId
	gotReport, err := db.GetUserReport(reportId)
	if err != nil {
		t.Errorf("test failed with err: %v", err)
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