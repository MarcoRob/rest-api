package user_report

import (
	"database/sql"
	"fmt"
	"database/sql/driver"
	"github.com/go-sql-driver/mysql"
	"encoding/json"
	"log"
)

const dbDoesNotExistError = 1049
const tableDoesNotExistError = 1146
const insertStatement = `
		INSERT INTO user_reports(user_id, today_tasks, delayed_tasks, 
			good_habits, bad_habits)
		VALUES (?, ?, ?, ?, ?);
	`
const getStatement = `
		SELECT report_id
		FROM user_reports
		WHERE user_id = ? AND today_tasks = ? AND delayed_tasks = ? 
			AND good_habits = ? AND bad_habits = ?;
	`
const getByIDStatement = `
		SELECT *
		FROM user_reports
		WHERE report_id = ?;
	`
var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS arqui DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE arqui;`,
	`CREATE TABLE IF NOT EXISTS user_reports (
		report_id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		today_tasks TEXT,
		delayed_tasks TEXT,
		good_habits TEXT,
		bad_habits TEXT
	);`,
}

type mysqlDB struct {
	conn *sql.DB

	insert 		*sql.Stmt
	getById 	*sql.Stmt
	get			*sql.Stmt
}

var _ UserReportDatabase = &mysqlDB{}

type MySQLConfig struct {
	Username, Password 	string
	Host 				string
	Port 				int
}

// return connection string for sql.Open
func (c MySQLConfig) dataStoreName(dbName string) string {
	var credentials string
	if c.Username != "" {
		credentials = c.Username
		if c.Password != "" {
			credentials = credentials + ":" + c.Password
		}
		credentials = credentials + "@"
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", credentials, c.Host, c.Port, dbName)
}

func newMySQLDB(config MySQLConfig) (UserReportDatabase, error) {
	if err := config.ensureTableExists(); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", config.dataStoreName("arqui"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &mysqlDB {
		conn: conn,
	}

	if db.getById, err = conn.Prepare(getByIDStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare getById: %v", err)
	}
	if db.get, err = conn.Prepare(getStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}
	if db.insert, err = conn.Prepare(insertStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert: %v", err)
	}

	return db, nil
}

func (db *mysqlDB) Close() {
	db.conn.Close()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanUserReport(s rowScanner) (*UserReport, error) {
	var (
		reportId		int64
		userId			string
		tasksToday		sql.NullString
		tasksDelayed 	sql.NullString
		habitsGood		sql.NullString
		habitsBad		sql.NullString
	)
	if err := s.Scan(&reportId, &userId, &tasksToday, &tasksDelayed, &habitsGood,
		&habitsBad); err != nil {
		return nil, err
	}

	var tasksTodaySlice, tasksDelayedSlice []Task
	var habitsGoodSlice, habitsBadSlice []Habit

	json.Unmarshal([]byte(tasksToday.String), &tasksTodaySlice)
	json.Unmarshal([]byte(tasksDelayed.String), &tasksDelayedSlice)
	json.Unmarshal([]byte(habitsGood.String), &habitsGoodSlice)
	json.Unmarshal([]byte(habitsBad.String), &habitsBadSlice)

	report := &UserReport{
		UserID:userId,
		ReportID:reportId,
		TodayTasks:tasksTodaySlice,
		DelayedTasks:tasksDelayedSlice,
		GoodHabits:habitsGoodSlice,
		BadHabits:habitsBadSlice,
	}
	return report, nil
}

// if table doesn't exist, create it
func (config MySQLConfig) ensureTableExists() error {
	conn, err := sql.Open("mysql", config.dataStoreName(""))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to db. ")
	}

	if _, err := conn.Exec(`USE arqui`); err != nil {
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == dbDoesNotExistError {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec(`DESCRIBE user_reports`); err != nil {
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == tableDoesNotExistError {
			return createTable(conn)
		}
		return fmt.Errorf("mysql: could not connect to db: %v", err)
	}
	return nil
}

// create db and table, as necessary
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execute a statement, expecting one row affected
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	result, err := stmt.Exec(args...)
	if err != nil {
		return result, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return result, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return result, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return result, nil
}

func (db *mysqlDB) GetUserReport(reportId int64) (*UserReport, error) {
	report, err := scanUserReport(db.getById.QueryRow(reportId))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("msql: could not find report with id %d", reportId)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get user report: %v", err)
	}
	return report, nil
}

func (db *mysqlDB) AddUserReport(report *UserReport) (reportId int64, err error) {
	if exists, reportId, err := db.checkExistingReport(report); err != nil {
		return -1, err
	} else if exists {
		log.Println("Report exists")
		return reportId, nil
	}

	todayTasksJSON, err := json.Marshal(report.TodayTasks)
	delayedTasksJSON, err := json.Marshal(report.DelayedTasks)
	goodHabitsJSON, err := json.Marshal(report.GoodHabits)
	badHabitsJSON, err := json.Marshal(report.BadHabits)

	result, err := execAffectingOneRow(db.insert, report.UserID, todayTasksJSON,
		delayedTasksJSON, goodHabitsJSON, badHabitsJSON)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}

	return lastInsertID, nil
}

func (db *mysqlDB) checkExistingReport(report *UserReport) (exists bool,
		reportId int64, err error) {
	todayTasksJSON, err := json.Marshal(report.TodayTasks)
	delayedTasksJSON, err := json.Marshal(report.DelayedTasks)
	goodHabitsJSON, err := json.Marshal(report.GoodHabits)
	badHabitsJSON, err := json.Marshal(report.BadHabits)

	err = db.get.QueryRow(report.UserID, todayTasksJSON,
		delayedTasksJSON, goodHabitsJSON, badHabitsJSON).Scan(&reportId)
	switch {
	case err == sql.ErrNoRows:
		return false, -1, nil
	case err != nil:
		log.Panic(err)
		return false, -1, err
	default:
		return true, reportId, nil
	}
}
