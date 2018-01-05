package course

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asepnur/meiko_information/src/util/helper"

	"database/sql"

	"github.com/asepnur/meiko_information/src/util/conn"
	"github.com/jmoiron/sqlx"
)

func IsEnrolled(userID, scheduleID int64) bool {

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "student")
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}

	if !res.Data.Involved {
		return false
	}

	return true
}

func IsAssistant(userID, scheduleID int64) bool {

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "assistant")
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}
	fmt.Println(res)
	if !res.Data.Involved {
		return false
	}

	return true
}

func IsExistSchedule(semester int8, year int16, courseID, class string, scheduleID ...int64) bool {

	var sc string
	if len(scheduleID) == 1 {
		sc = fmt.Sprintf(" AND id != (%d) ", scheduleID[0])
	}

	var x string
	query := fmt.Sprintf(`
		SELECT
			'x'
		FROM
			schedules
		WHERE
			semester = (%d) AND
			year = (%d) AND
			courses_id = ('%s') AND
			class = ('%s') %s
		LIMIT 1;`, semester, year, courseID, class, sc)
	err := conn.DB.Get(&x, query)
	if err != nil {
		return false
	}
	return true
}

func InsertSchedule(userID int64, startTime, endTime, year int16, semester, day, status int8, class, courseID, placeID string, tx ...*sqlx.Tx) (int64, error) {

	var id int64
	query := fmt.Sprintf(`
		INSERT INTO
			schedules (
				status,
				start_time,
				end_time,
				day,
				class,
				semester,
				year,
				courses_id,
				places_id,
				created_by,
				created_at,
				updated_at
			)
		VALUES (
			(%d),
			(%d),
			(%d),
			(%d),
			('%s'),
			(%d),
			(%d),
			('%s'),
			('%s'),
			(%d),
			NOW(),
			NOW()
		)`, status, startTime, endTime, day, class, semester, year, courseID, placeID, userID)

	var result sql.Result
	var err error
	switch len(tx) {
	case 1:
		result, err = tx[0].Exec(query)
	default:
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return id, err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return id, fmt.Errorf("Cannot get last insert id")
	}

	return id, nil
}

func SelectByPage(limit, offset int, isCount bool) ([]CourseSchedule, int, error) {

	var course []CourseSchedule
	var count int
	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.name,
			cs.description,
			cs.ucu,
			sc.id,
			sc.status,
			sc.start_time,
			sc.end_time,
			sc.day,
			sc.class,
			sc.semester,
			sc.year,
			sc.places_id,
			sc.created_by
		FROM
			courses cs
		RIGHT JOIN
			schedules sc
		ON
			cs.id = sc.courses_id
		LIMIT %d OFFSET %d;`, limit, offset)
	rows, err := conn.DB.Queryx(query)
	defer rows.Close()
	if err != nil {
		return course, count, err
	}

	for rows.Next() {
		var id, name, class, placeID string
		var description sql.NullString
		var ucu, status, day, semester int8
		var startTime, endTime uint16
		var year int16
		var scheduleID, createdBy int64

		err := rows.Scan(&id, &name, &description, &ucu, &scheduleID, &status, &startTime, &endTime, &day, &class, &semester, &year, &placeID, &createdBy)
		if err != nil {
			return course, count, err
		}

		course = append(course, CourseSchedule{
			Course: Course{
				ID:          id,
				Name:        name,
				Description: description,
				UCU:         ucu,
			},
			Schedule: Schedule{
				ID:        scheduleID,
				Status:    status,
				StartTime: startTime,
				EndTime:   endTime,
				Day:       day,
				Class:     class,
				Semester:  semester,
				Year:      year,
				PlaceID:   placeID,
				CreatedBy: createdBy,
			},
		})
	}

	if !isCount {
		return course, count, nil
	}

	query = fmt.Sprintf(`
		SELECT
			COUNT(*)
		FROM
			schedules`)
	err = conn.DB.Get(&count, query)
	if err != nil {
		return course, count, err
	}

	return course, count, nil
}

func GetByScheduleID(scheduleID int64) (CourseSchedule, error) {

	var course CourseSchedule
	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.name,
			cs.description,
			cs.ucu,
			sc.id,
			sc.status,
			sc.start_time,
			sc.end_time,
			sc.day,
			sc.class,
			sc.semester,
			sc.year,
			sc.places_id,
			sc.created_by
		FROM
			courses cs
		RIGHT JOIN
			schedules sc
		ON
			cs.id = sc.courses_id
		WHERE
			sc.id = (%d)
		LIMIT 1;`, scheduleID)

	rows := conn.DB.QueryRowx(query)

	// scan data to variable
	var id, name, class, placeID string
	var description sql.NullString
	var ucu, status, day, semester int8
	var startTime, endTime uint16
	var year int16
	var createdBy int64

	err := rows.Scan(&id, &name, &description, &ucu, &scheduleID, &status, &startTime, &endTime, &day, &class, &semester, &year, &placeID, &createdBy)
	if err != nil {
		return course, err
	}

	return CourseSchedule{
		Course: Course{
			ID:          id,
			Name:        name,
			Description: description,
			UCU:         ucu,
		},
		Schedule: Schedule{
			ID:        scheduleID,
			Status:    status,
			StartTime: startTime,
			EndTime:   endTime,
			Day:       day,
			Class:     class,
			Semester:  semester,
			Year:      year,
			PlaceID:   placeID,
			CreatedBy: createdBy,
		},
	}, nil
}

func IsExistScheduleID(scheduleID int64) bool {
	var x string
	query := fmt.Sprintf(`
		SELECT
			'x'
		FROM
			schedules
		WHERE
			id = (%d)
		LIMIT 1;`, scheduleID)
	err := conn.DB.Get(&x, query)
	if err != nil {
		return false
	}
	return true
}

func UpdateSchedule(scheduleID int64, startTime, endTime, year int16, semester, day, status int8, class, courseID, placeID string, tx ...*sqlx.Tx) error {

	query := fmt.Sprintf(`
		UPDATE 
			schedules
		SET
			status = (%d),
			start_time = (%d),
			end_time = (%d),
			day = (%d),
			class = ('%s'),
			semester = (%d),
			year = (%d),
			courses_id = ('%s'),
			places_id = ('%s'),
			updated_at = NOW()
		WHERE
			id = (%d);`, status, startTime, endTime, day, class, semester, year, courseID, placeID, scheduleID)

	var result sql.Result
	var err error
	switch len(tx) {
	case 1:
		result, err = tx[0].Exec(query)
	default:
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

func SelectByScheduleID(scheduleID []int64, status int8) ([]CourseSchedule, error) {

	var course []CourseSchedule
	if len(scheduleID) < 1 {
		return course, nil
	}

	d := helper.Int64ToStringSlice(scheduleID)
	ids := strings.Join(d, ", ")

	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.name,
			cs.description,
			cs.ucu,
			sc.id,
			sc.status,
			sc.start_time,
			sc.end_time,
			sc.day,
			sc.class,
			sc.semester,
			sc.year,
			sc.places_id,
			sc.created_by
		FROM
			courses cs
		RIGHT JOIN
			schedules sc
		ON
			cs.id = sc.courses_id
		WHERE
			sc.id IN (%s) AND
			sc.status = (%d)
		ORDER BY day ASC;`, ids, status)

	rows, err := conn.DB.Queryx(query)
	defer rows.Close()
	if err != nil {
		return course, err
	}

	for rows.Next() {
		var id, name, class, placeID string
		var description sql.NullString
		var ucu, status, day, semester int8
		var startTime, endTime uint16
		var year int16
		var scID, createdBy int64

		err := rows.Scan(&id, &name, &description, &ucu, &scID, &status, &startTime, &endTime, &day, &class, &semester, &year, &placeID, &createdBy)
		if err != nil {
			return course, err
		}

		course = append(course, CourseSchedule{
			Course: Course{
				ID:          id,
				Name:        name,
				Description: description,
				UCU:         ucu,
			},
			Schedule: Schedule{
				ID:        scID,
				Status:    status,
				StartTime: startTime,
				EndTime:   endTime,
				Day:       day,
				Class:     class,
				Semester:  semester,
				Year:      year,
				PlaceID:   placeID,
				CreatedBy: createdBy,
			},
		})
	}

	return course, nil
}

func SelectByStatus(status int8) ([]CourseSchedule, error) {

	var course []CourseSchedule
	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.name,
			cs.description,
			cs.ucu,
			sc.id,
			sc.status,
			sc.start_time,
			sc.end_time,
			sc.day,
			sc.class,
			sc.semester,
			sc.year,
			sc.places_id,
			sc.created_by
		FROM
			courses cs
		RIGHT JOIN
			schedules sc
		ON
			cs.id = sc.courses_id
		WHERE
			sc.status = (%d)`, status)

	rows, err := conn.DB.Queryx(query)
	defer rows.Close()
	if err != nil {
		return course, err
	}

	for rows.Next() {
		var id, name, class, placeID string
		var description sql.NullString
		var ucu, status, day, semester int8
		var startTime, endTime uint16
		var year int16
		var scheduleID, createdBy int64

		err := rows.Scan(&id, &name, &description, &ucu, &scheduleID, &status, &startTime, &endTime, &day, &class, &semester, &year, &placeID, &createdBy)
		if err != nil {
			return course, err
		}

		course = append(course, CourseSchedule{
			Course: Course{
				ID:          id,
				Name:        name,
				Description: description,
				UCU:         ucu,
			},
			Schedule: Schedule{
				ID:        scheduleID,
				Status:    status,
				StartTime: startTime,
				EndTime:   endTime,
				Day:       day,
				Class:     class,
				Semester:  semester,
				Year:      year,
				PlaceID:   placeID,
				CreatedBy: createdBy,
			},
		})
	}

	return course, nil
}

func DeleteSchedule(scheduleID int64, tx ...*sqlx.Tx) error {

	query := fmt.Sprintf(`
		DELETE FROM
			schedules
		WHERE
			id = (%d);
		`, scheduleID)

	var result sql.Result
	var err error
	if len(tx) == 1 {
		result, err = tx[0].Exec(query)
	} else {
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

func SelectByName(name string) ([]Course, error) {
	var courses []Course
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			description,
			ucu
		FROM
			courses
		WHERE
			name LIKE ('%%%s%%')
		LIMIT 5;
	`, name)
	err := conn.DB.Select(&courses, query)
	if err != nil && err != sql.ErrNoRows {
		return courses, err
	}

	return courses, nil
}

func InsertGradeParameter(typ string, percentage float32, statusChange uint8, scheduleID int64, tx *sqlx.Tx) error {

	query := fmt.Sprintf(`
		INSERT INTO
		grade_parameters (
			type,
			percentage,
			status_change,
			schedules_id,
			created_at,
			updated_at
		)
		VALUES (
			('%s'),
			(%f),
			(%d),
			(%d),
			NOW(),
			NOW()
		);
		`, typ, percentage, statusChange, scheduleID)

	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.Exec(query)
	} else {
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

func SelectGPBySchedule(scheduleID []int64) ([]GradeParameter, error) {
	var gps []GradeParameter

	if len(scheduleID) < 1 {
		return gps, nil
	}

	querySchID := strings.Join(helper.Int64ToStringSlice(scheduleID), ", ")
	query := fmt.Sprintf(`
		SELECT
			id,
			type,
			percentage,
			status_change,
			schedules_id
		FROM
			grade_parameters
		WHERE
			schedules_id IN (%s);
		`, querySchID)
	err := conn.DB.Select(&gps, query)
	if err != nil && err != sql.ErrNoRows {
		return gps, err
	}
	return gps, nil
}

// SelectGradeParameterByScheduleIDIN func
func SelectGradeParameterByScheduleIDIN(scheduleID []int64) ([]int64, error) {
	var gps []int64
	var gradeQuery []string
	for _, value := range scheduleID {
		gradeQuery = append(gradeQuery, fmt.Sprintf("%d", value))
	}
	queryGradeList := strings.Join(gradeQuery, ",")
	query := fmt.Sprintf(`
		SELECT
			id
		FROM
			grade_parameters
		WHERE
			schedules_id
		IN
			 (%s);
		`, queryGradeList)
	err := conn.DB.Select(&gps, query)
	if err != nil && err != sql.ErrNoRows {
		return gps, err
	}
	return gps, nil
}

func DeleteGradeParameter(id int64, tx *sqlx.Tx) error {

	query := fmt.Sprintf(`
			DELETE FROM
				grade_parameters
			WHERE
				id = (%d);
			`, id)

	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.Exec(query)
	} else {
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

func UpdateGradeParameter(typ string, percentage float32, statusChange uint8, scheduleID int64, tx *sqlx.Tx) error {

	query := fmt.Sprintf(`
		UPDATE
			grade_parameters
		SET
			percentage = (%f),
			status_change = (%d),
			updated_at = NOW()
		WHERE
			type = ('%s') AND
			schedules_id = (%d);
		`, percentage, statusChange, typ, scheduleID)

	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.Exec(query)
	} else {
		result, err = conn.DB.Exec(query)
	}
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("No rows affected")
	}

	return nil
}

// GetScheduleIDByGP ...
func GetScheduleIDByGP(gpID int64) (int64, error) {
	var scheduleID int64
	query := fmt.Sprintf(`
		SELECT 
			schedules_id
		FROM
			grade_parameters
		WHERE
			id = (%d)
		`, gpID)
	err := conn.DB.Get(&scheduleID, query)
	if err != nil {
		return scheduleID, err
	}

	return scheduleID, nil
}

// GetGradeParametersID func ...
func GetGradeParametersID(AssignmentID int64) int64 {
	query := fmt.Sprintf(`
		SELECT 
			asg.grade_parameters_id
		FROM
			assignments asg
		WHERE
			id = (%d)
		`, AssignmentID)
	var assignmentID string
	err := conn.DB.QueryRow(query).Scan(&assignmentID)
	if err != nil {
		return 0
	}
	assignmentid, err := strconv.ParseInt(assignmentID, 10, 64)
	if err != nil {
		return 0
	}

	return assignmentid
}

// GetGradeParametersIDByScheduleID func ...
func GetGradeParametersIDByScheduleID(ScheduleID int64) ([]int64, error) {
	query := fmt.Sprintf(`
		SELECT
			id
		FROM
			grade_parameters
		WHERE
			schedules_id = (%d)
		;`, ScheduleID)

	rows, err := conn.DB.Query(query)
	var gradeParamsID []int64
	if err != nil {
		return gradeParamsID, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return gradeParamsID, err
		}
		gradeParamsID = append(gradeParamsID, id)
	}
	return gradeParamsID, nil
}

// IsUserHasUploadedFile func ...
func IsUserHasUploadedFile(assignmentID, userID int64) bool {
	var x string
	query := fmt.Sprintf(`
		SELECT
			'x'
		FROM
			p_users_assignments
		WHERE
			assignments_id = (%d) AND users_id =(%d)
		LIMIT 1;`, assignmentID, userID)
	err := conn.DB.Get(&x, query)
	if err != nil {
		return false
	}
	return true
}

// IsAllUsersEnrolled func ...
func IsAllUsersEnrolled(scheduleID int64, usersID []int64) bool {
	userIDs := []int64{}
	var userList []string
	for _, value := range usersID {
		userList = append(userList, fmt.Sprintf("%d", value))
	}
	queryUserList := strings.Join(userList, ",")
	query := fmt.Sprintf(`SELECT
			users_id
		FROM
			p_users_schedules
		WHERE 
			status = (%d) AND
			schedules_id = (%d) AND users_id IN(%s);`, PStatusStudent, scheduleID, queryUserList)

	err := conn.DB.Select(&userIDs, query)
	if err != nil && err != sql.ErrNoRows {
		return false
	}

	if len(userIDs) != len(usersID) {
		return false
	}
	return true
}

// GetCourseID func ...
func GetCourseID(scheduleID int64) (string, error) {
	query := fmt.Sprintf(`
		SELECT
			courses_id
		FROM
			schedules
		WHERE
			id=(%d)
		LIMIT 1;
		`, scheduleID)
	var res string
	err := conn.DB.Get(&res, query)
	if err != nil {
		return res, err
	}
	return res, nil
}

// GetName func ...
func GetName(courseID string) (string, error) {
	query := fmt.Sprintf(`
		SELECT
			name
		FROM
			courses
		WHERE
			id=('%s')
		LIMIT 1;
		`, courseID)
	var res string
	err := conn.DB.Get(&res, query)
	if err != nil {
		return res, err
	}
	return res, nil
}
func SelectJoinScheduleCourse(scheduleID []int64) ([]CourseConcise, error) {
	var res []CourseConcise
	id := helper.Int64ToStringSlice(scheduleID)
	queryID := strings.Join(id, ", ")
	query := fmt.Sprintf(`
		SELECT
			sc.id,
			cs.name
		FROM
			schedules sc
		INNER JOIN
			courses cs
		ON
			sc.courses_id = cs.id
		WHERE 
			sc.id IN (%s)
		;
		`, queryID)
	err := conn.DB.Select(&res, query)
	if err != nil {
		return res, err
	}
	return res, nil

}

// InsertAssistant ...
func InsertAssistant(usersID []int64, scheduleID int64, tx *sqlx.Tx) error {

	var values []string
	for _, val := range usersID {
		value := fmt.Sprintf("(%d, %d, %d, NOW(), NOW())", val, scheduleID, PStatusAssistant)
		values = append(values, value)
	}

	queryValue := strings.Join(values, ", ")
	query := fmt.Sprintf(`
		INSERT INTO
			p_users_schedules (
				users_id,
				schedules_id,
				status,
				created_at,
				updated_at
			) VALUES %s;
	`, queryValue)

	var err error
	if tx != nil {
		_, err = tx.Exec(query)
	} else {
		_, err = conn.DB.Exec(query)
	}

	if err != nil {
		return err
	}
	return nil
}

func DeleteAssistant(usersID []int64, scheduleID int64, tx *sqlx.Tx) error {

	usersIDString := helper.Int64ToStringSlice(usersID)
	queryUsersID := strings.Join(usersIDString, ", ")
	query := fmt.Sprintf(`
		DELETE FROM
			p_users_schedules
		WHERE
			status = (%d) AND
			schedules_id = (%d) AND
			users_id IN (%s);
	`, PStatusAssistant, scheduleID, queryUsersID)

	var err error
	if tx != nil {
		_, err = tx.Exec(query)
	} else {
		_, err = conn.DB.Exec(query)
	}

	if err != nil {
		return err
	}
	return nil
}

func SelectByDayScheduleID(day int8, schedulesID []int64) ([]CourseSchedule, error) {
	var course []CourseSchedule
	if len(schedulesID) < 1 {
		return course, nil
	}

	querySchedulesID := strings.Join(helper.Int64ToStringSlice(schedulesID), ", ")
	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.name,
			cs.description,
			cs.ucu,
			sc.id,
			sc.status,
			sc.start_time,
			sc.end_time,
			sc.day,
			sc.class,
			sc.semester,
			sc.year,
			sc.places_id,
			sc.created_by
		FROM
			courses cs
		RIGHT JOIN
			schedules sc
		ON
			cs.id = sc.courses_id
		WHERE
			sc.day = (%d) AND
			sc.id IN (%s)
		;`, day, querySchedulesID)
	rows, err := conn.DB.Queryx(query)
	defer rows.Close()
	if err != nil {
		return course, err
	}

	for rows.Next() {
		var id, name, class, placeID string
		var description sql.NullString
		var ucu, status, day, semester int8
		var startTime, endTime uint16
		var year int16
		var scheduleID, createdBy int64

		err := rows.Scan(&id, &name, &description, &ucu, &scheduleID, &status, &startTime, &endTime, &day, &class, &semester, &year, &placeID, &createdBy)
		if err != nil {
			return course, err
		}

		course = append(course, CourseSchedule{
			Course: Course{
				ID:          id,
				Name:        name,
				Description: description,
				UCU:         ucu,
			},
			Schedule: Schedule{
				ID:        scheduleID,
				Status:    status,
				StartTime: startTime,
				EndTime:   endTime,
				Day:       day,
				Class:     class,
				Semester:  semester,
				Year:      year,
				PlaceID:   placeID,
				CreatedBy: createdBy,
			},
		})
	}

	return course, nil
}

func InsertUnapproved(userID, scheduleID int64) error {
	query := fmt.Sprintf(`
		INSERT INTO
			p_users_schedules (
				users_id,
				schedules_id,
				status,
				created_at,
				updated_at
			) VALUES (
				(%d),
				(%d),
				(%d),
				NOW(),
				NOW()
			);
	`, userID, scheduleID, PStatusUnapproved)

	_, err := conn.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUserRelation(userID, scheduleID int64) error {

	query := fmt.Sprintf(`
		DELETE FROM
			p_users_schedules
		WHERE
			users_id = (%d) AND
			schedules_id = (%d);
	`, userID, scheduleID)

	result, err := conn.DB.Exec(query)
	if err != nil {
		return err
	}

	if valid, err := result.RowsAffected(); err != nil || valid < 1 {
		return err
	}

	return nil
}

func SelectEnrolledSchedule(userID int64) ([]CourseSchedule, error) {

	var res []CourseSchedule

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "student")

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getall", strings.NewReader(params))
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	jsn := GetAllHTTPResponse{}
	err = json.Unmarshal(body, &jsn)
	if err != nil {
		return res, err
	}

	if jsn.Code != 200 {
		return res, fmt.Errorf("error request")
	}

	return jsn.Data, nil
}

func SelectTeachedSchedule(userID int64) ([]CourseSchedule, error) {

	var res []CourseSchedule

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "assistant")

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getall", strings.NewReader(params))
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	jsn := GetAllHTTPResponse{}
	err = json.Unmarshal(body, &jsn)
	if err != nil {
		return res, err
	}

	if jsn.Code != 200 {
		return res, fmt.Errorf("error request")
	}

	return jsn.Data, nil
}

func Get(userID, scheduleID int64, isAssistant bool) (GetOne, error) {

	var getOne GetOne

	role := "student"
	if isAssistant {
		role = "assistant"
	}

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", role)
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return getOne, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return getOne, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return getOne, err
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return getOne, err
	}

	return res.Data, nil
}
