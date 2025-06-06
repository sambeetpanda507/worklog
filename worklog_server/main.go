package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// structure of the secrets
type secrets struct {
	host       string
	dbPort     string
	dbUsername string
	dbPassword string
	dbName     string
	serverPort string
}

type WorkLog struct {
	LogId       string     `json:"logId"`
	TaskName    string     `json:"taskName"`
	TaskType    string     `json:"taskType"`
	TaskStatus  string     `json:"taskStatus"`
	Notes       string     `json:"notes"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Priority    int        `json:"priority"`
}

// get the secrets
func GetSecrets() secrets {
	secrets := secrets{
		host:       os.Getenv("DB_HOST"),
		dbPort:     os.Getenv("DB_PORT"),
		dbUsername: os.Getenv("DB_USERNAME"),
		dbPassword: os.Getenv("DB_PASSWORD"),
		dbName:     os.Getenv("DB_NAME"),
		serverPort: os.Getenv("SERVER_PORT"),
	}

	return secrets
}

// get the database connection string
func GetConnectionStr() string {
	secrets := GetSecrets()
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		secrets.host, secrets.dbPort, secrets.dbUsername, secrets.dbPassword, secrets.dbName,
	)

	return connectionString
}

// connect to database
func ConnectToDB() *sql.DB {
	connectionStr := GetConnectionStr()
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		panic(pingErr)
	}

	log.Println("Database connected successfully")
	return db
}

// get all the logs
type GetAllLogsOpts struct {
	db        *sql.DB
	s         string
	sortBy    string
	sortOrder string
	limit     int16
	page      int16
}

func getFullTextSearchQueryOnLogs(userQuery string, limit int16, offset int16) string {
	tsQueryText := strings.Replace(userQuery, " ", " & ", -1)
	fmt.Println("tsQueryText: ", tsQueryText)
	// q := fmt.Sprintf(`
	// 	select
	// 		log_id,
	// 		task_name,
	// 		task_type,
	// 		task_status,
	// 		notes,
	// 		started_at,
	// 		completed_at,
	// 		created_at,
	// 		updated_at,
	// 		priority,
	// 		(ceil((select count(*) from logs) / %d)) as total_pages
	// 	from logs
	// 	where ts @@ to_tsquery('english', '%s')
	// 	or similarity('%s', task_name || ' ' || notes) > 0
	// 	order by
	// 		ts_rank(ts, to_tsquery('english', '%s')) desc,
	// 		similarity('%s', task_name || ' ' || notes) desc
	// 	offset %d limit %d`,
	// 	limit,
	// 	tsQueryText,
	// 	userQuery,
	// 	tsQueryText,
	// 	userQuery,
	// 	offset,
	// 	limit,
	// )

	q := fmt.Sprintf(`
		with filtered_logs as (
			select * from logs
			where ts @@ to_tsquery('english', '%s')
			or similarity('%s', task_name || ' ' || notes) > 0
		) 
	 	select
	 		log_id,
	 		task_name,
	 		task_type,
	 		task_status,
	 		notes,
	 		started_at,
	 		completed_at,
	 		created_at,
	 		updated_at,
	 		priority,
	 		(ceil(count(*) over() / %f)) as total_pages
		from filtered_logs
		order by
			ts_rank(ts, to_tsquery('english', '%s')) desc,
			similarity('%s', task_name || ' ' || notes) desc
		offset %d limit %d`,
		tsQueryText,
		userQuery,
		float64(limit),
		tsQueryText,
		userQuery,
		offset,
		limit,
	)

	return q
}

func GetAllLogs(options *GetAllLogsOpts) (*sql.Rows, error) {
	var q string
	var sortBy string
	var sortOrder string
	var limit int16
	var offset int16

	if options.sortBy == "" {
		sortBy = "updated_at"
	} else {
		sortBy = options.sortBy
	}

	if options.sortOrder == "" {
		sortOrder = "desc"
	} else {
		sortOrder = options.sortOrder
	}

	if options.limit > 0 {
		limit = options.limit
	} else {
		limit = 10
	}

	if options.page > 0 {
		offset = options.page * limit
	} else {
		offset = 0
	}

	// check if options have search value
	if strings.TrimSpace(options.s) != "" {
		q = getFullTextSearchQueryOnLogs(options.s, limit, offset)
		// q = fmt.Sprintf(
		// 	`select * from logs
		// 	where task_name ilike '%%%s%%'
		// 	or notes ilike '%%%s%%'
		// 	order by updated_at desc;`,
		// 	options.s,
		// 	options.s,
		// )
	} else {
		q = fmt.Sprintf(`select
			log_id,
			task_name,
			task_type,
			task_status,
			notes,
			started_at,
			completed_at,
			created_at,
			updated_at,
			priority,
			(ceil((select count(*) from logs) / %f)) as total_pages 
		from logs order by %s %s 
		offset %d limit %d;`, float64(limit), sortBy, sortOrder, offset, limit)
	}

	fmt.Println("[query]: ", q)
	rows, err := options.db.Query(q)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func handleCreateLog(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		TaskName    string     `json:"taskName"`
		TaskType    string     `json:"taskType"`
		TaskStatus  string     `json:"taskStatus"`
		Notes       string     `json:"notes"`
		StartedAt   *time.Time `json:"startedAt"`
		CompletedAt *time.Time `json:"completedAt"`
		Priority    *int       `json:"priority"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.TrimSpace(body.TaskName) == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "Task name is required"})
		return
	}

	if body.TaskStatus == "" || !validateTaskStatus(body.TaskStatus) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid task status"})
		return
	}

	if body.TaskType == "" || !validateTaskType(body.TaskType) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid task type"})
		return
	}

	if body.Notes == "" {
		body.Notes = "N/A"
	}

	if body.StartedAt != nil {
		startedAt := body.StartedAt.UTC()
		body.StartedAt = &startedAt
	}

	if body.CompletedAt != nil {
		completedAt := body.CompletedAt.UTC()
		body.CompletedAt = &completedAt
	}

	// handle task priority
	if body.Priority != nil {
		// valid priority int
		if *body.Priority != 1 && *body.Priority != 5 && *body.Priority != 7 && *body.Priority != 10 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid priority value"})
			return
		}
	}

	// assing default value of priority if it's nil
	if body.Priority == nil {
		var defaultVal int = 1
		body.Priority = &defaultVal
	}

	// check for duplicate keys
	q := "select 1 from logs where task_name = $1 limit 1"
	row := db.QueryRow(q, body.TaskName)
	var exists int
	err = row.Scan(&exists)
	if err == sql.ErrNoRows {
		// no duplicate found, proceed
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// duplicate found
		http.Error(w, "Task name already exists", http.StatusBadRequest)
		return
	}

	q = "insert into logs (task_name, task_type, task_status, notes, started_at, completed_at, priority) values ($1, $2, $3, $4, $5, $6, $7)"
	_, err = db.Exec(
		q,
		body.TaskName,
		body.TaskType,
		body.TaskStatus,
		body.Notes,
		body.StartedAt,
		body.CompletedAt,
		body.Priority,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateTaskType(taskType string) bool {
	taskTypes := map[string]bool{
		"task":  true,
		"bug":   true,
		"story": true,
	}

	return taskTypes[taskType]
}

func validateTaskStatus(taskStatus string) bool {
	taskTypes := map[string]bool{
		"backlog":  true,
		"pending":  true,
		"progress": true,
		"pr":       true,
		"staging":  true,
	}

	return taskTypes[taskStatus]
}

func getLogById(db *sql.DB, logId string) (WorkLog, error) {
	q := "select log_id, task_name, task_type, task_status, priority, notes, started_at, completed_at, created_at, updated_at from logs where log_id = $1"
	row := db.QueryRow(q, logId)
	var workLog WorkLog
	var (
		notes       sql.NullString
		startedAt   sql.NullTime
		completedAt sql.NullTime
	)

	err := row.Scan(
		&workLog.LogId,
		&workLog.TaskName,
		&workLog.TaskType,
		&workLog.TaskStatus,
		&workLog.Priority,
		&notes,
		&startedAt,
		&completedAt,
		&workLog.CreatedAt,
		&workLog.UpdatedAt,
	)

	if err != nil {
		return workLog, err
	}

	if notes.Valid {
		workLog.Notes = notes.String
	} else {
		workLog.Notes = "n/a"
	}

	if startedAt.Valid {
		workLog.StartedAt = &startedAt.Time
	} else {
		workLog.StartedAt = nil
	}

	if completedAt.Valid {
		workLog.CompletedAt = &completedAt.Time
	} else {
		workLog.CompletedAt = nil
	}

	return workLog, nil
}

func updateLog(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json")

	// get the body from request
	var body struct {
		LogId       string     `json:"logId"`
		TaskName    string     `json:"taskName"`
		TaskType    string     `json:"taskType"`
		TaskStatus  string     `json:"taskStatus"`
		Notes       string     `json:"notes"`
		StartedAt   *time.Time `json:"startedAt"`
		CompletedAt *time.Time `json:"completedAt"`
		Priority    *int       `json:"priority"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if body.LogId == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Log id is required",
		})
		return
	}

	q := "select log_id from logs where log_id = $1"
	row := db.QueryRow(q, body.LogId)
	var existingData struct {
		LogId string `json:"log_id"`
	}

	err = row.Scan(&existingData.LogId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No records found with this log id",
				"logId":   body.LogId,
			})

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(struct {
			Message string `json:"message"`
		}{Message: err.Error()})

		return
	}

	fields := []string{}
	args := []any{}
	argIdx := 1

	if strings.TrimSpace(body.TaskName) != "" {
		fields = append(fields, fmt.Sprintf("task_name = $%d", argIdx))
		args = append(args, body.TaskName)
		argIdx++
	}

	if body.TaskType != "" {
		// validate task type
		if !validateTaskType(body.TaskType) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid task type"})
			return
		}

		fields = append(fields, fmt.Sprintf("task_type = $%d", argIdx))
		args = append(args, body.TaskType)
		argIdx++
	}

	if body.TaskStatus != "" {
		// validate task status
		if !validateTaskStatus(body.TaskStatus) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid task status"})
			return
		}

		fields = append(fields, fmt.Sprintf("task_status = $%d", argIdx))
		args = append(args, body.TaskStatus)
		argIdx++
	}

	if strings.TrimSpace(body.Notes) != "" {
		fields = append(fields, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, body.Notes)
		argIdx++
	}

	if body.StartedAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, body.StartedAt.Format(time.RFC3339))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing started at value"})
			return
		}

		fields = append(fields, fmt.Sprintf("started_at = $%d", argIdx))
		args = append(args, parsedTime.UTC())
		argIdx++
	}

	if body.CompletedAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, body.CompletedAt.Format(time.RFC3339))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Error parsing completed at value"})
			return
		}

		fields = append(fields, fmt.Sprintf("completed_at = $%d", argIdx))
		args = append(args, parsedTime.UTC())
		argIdx++
	}

	if body.Priority != nil {
		if *body.Priority != 1 && *body.Priority != 5 && *body.Priority != 7 && *body.Priority != 10 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid priority value"})
			return
		}

		fields = append(fields, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *body.Priority)
		argIdx += 1
	}

	if len(fields) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	args = append(args, body.LogId)
	query := fmt.Sprintf("update logs set %s, updated_at = now() where log_id = $%d", strings.Join(fields, ", "), argIdx)
	fmt.Println(query)

	if _, err := db.Exec(query, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	updatedLog, err := getLogById(db, body.LogId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Message string `json:"message"`
		Log     any    `json:"log"`
	}{
		Message: "Log updated successfully",
		Log:     updatedLog,
	}

	json.NewEncoder(w).Encode(response)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("origin: ", r.Header.Get("origin"))
		allowedOrigins := map[string]bool{
			"http://localhost:4173": true,
			"http://localhost:5173": true,
		}

		origin := r.Header.Get("origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db := ConnectToDB()
	defer db.Close()
	serverPort := GetSecrets().serverPort
	mux := http.NewServeMux()
	log.Println("server is running on http://localhost:" + serverPort)
	mux.HandleFunc("/ping", pingHandler)

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// get the search params
		searchVal := r.URL.Query().Get("s")
		sortBy := r.URL.Query().Get("sortBy")
		sortOrder := r.URL.Query().Get("sortOrder")
		limitStr := r.URL.Query().Get("limit")
		pageStr := r.URL.Query().Get("page")

		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid limit value", http.StatusUnprocessableEntity)
			return
		}

		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid page value", http.StatusUnprocessableEntity)
		}

		// lexical search
		// for task_name, task_type, task_status, notes
		rows, err := GetAllLogs(&GetAllLogsOpts{db: db, s: searchVal, sortBy: sortBy, sortOrder: sortOrder, limit: int16(limit), page: int16(page)})
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer rows.Close()
		type Log struct {
			LogId       string     `json:"logId"`
			TaskName    string     `json:"taskName"`
			TaskType    string     `json:"taskType"`
			TaskStatus  string     `json:"taskStatus"`
			Notes       string     `json:"notes"`
			StartedAt   *time.Time `json:"startedAt"`
			CompletedAt *time.Time `json:"completedAt"`
			CreatedAt   time.Time  `json:"createdAt"`
			UpdatedAt   time.Time  `json:"updatedAt"`
			Priority    int        `json:"priority"`
			TotalPages  int        `json:"totalPages"`
		}

		var logs []Log
		for rows.Next() {
			var (
				logEntry    Log
				notes       sql.NullString
				startedAt   sql.NullTime
				completedAt sql.NullTime
			)

			err := rows.Scan(
				&logEntry.LogId,
				&logEntry.TaskName,
				&logEntry.TaskType,
				&logEntry.TaskStatus,
				&notes,
				&startedAt,
				&completedAt,
				&logEntry.CreatedAt,
				&logEntry.UpdatedAt,
				&logEntry.Priority,
				&logEntry.TotalPages,
			)

			if err != nil {
				http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			}

			if notes.Valid {
				logEntry.Notes = notes.String
			} else {
				logEntry.Notes = "n/a"
			}

			if completedAt.Valid {
				logEntry.CompletedAt = &completedAt.Time
			} else {
				logEntry.CompletedAt = nil
			}

			if startedAt.Valid {
				logEntry.StartedAt = &startedAt.Time
			} else {
				logEntry.StartedAt = nil
			}

			logs = append(logs, logEntry)
		}

		if logs == nil {
			json.NewEncoder(w).Encode(struct {
				Logs []Log `json:"logs"`
			}{Logs: []Log{}})
			return
		}

		json.NewEncoder(w).Encode(struct {
			Logs []Log `json:"logs"`
		}{Logs: logs})
	})

	mux.HandleFunc("/log/{logId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logId := r.PathValue("logId")
		worklog, err := getLogById(db, logId)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"message": "No records found"})
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't wrong"})
			return
		}

		w.WriteHeader(http.StatusOK)
		type response struct {
			Message string  `json:"message"`
			Log     WorkLog `json:"log"`
		}

		json.NewEncoder(w).Encode(&response{Message: "Ok", Log: worklog})
	})

	mux.HandleFunc("PUT /log", func(w http.ResponseWriter, r *http.Request) {
		updateLog(w, r, db)
	})

	mux.HandleFunc("POST /log", func(w http.ResponseWriter, r *http.Request) {
		handleCreateLog(db, w, r)
	})

	mux.HandleFunc("DELETE /log/{logId}", func(w http.ResponseWriter, r *http.Request) {
		logId := r.PathValue("logId")
		w.Header().Set("Content-Type", "application/json")

		// validate log id
		q := "select log_id from logs where log_id = $1"
		var id string
		err := db.QueryRow(q, logId).Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"message": "No log found"})
				return
			}

			json.NewEncoder(w).Encode(struct {
				Message string `json:"message"`
			}{Message: err.Error()})
			return
		}

		// delete the log by log id
		q = "delete from logs where log_id = $1"
		_, err = db.Exec(q, logId)
		if err != nil {
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message"`
			}{Message: err.Error()})
			return
		}

		// return final response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Successfully delete the log"})
	})

	mux.HandleFunc("DELETE /logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logIds := r.URL.Query().Get("logIds")
		if logIds == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"message": "At least 1 log id is required."})
			return
		}

		var ids []string
		err := json.NewDecoder(strings.NewReader(logIds)).Decode(&ids)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid logIds format"})
			return
		}

		// Process the decoded log IDs
		// delete from logs where log_id in (a, b, c, d, ..., n)
		for i, id := range ids {
			ids[i] = fmt.Sprintf("'%s'", id)
		}

		q := fmt.Sprintf("delete from logs where log_id in (%s)", strings.Join(ids, ", "))
		result, err := db.Exec(q)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't wrong while deleting logs."})
			return
		}

		rowCount, err := result.RowsAffected()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't wrong while deleting logs."})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Message  string `json:"message"`
			RowCount int64  `json:"rowCount"`
		}{Message: "Logs deleted successfully", RowCount: rowCount})
	})

	mux.HandleFunc("GET /status-summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := `
			SELECT
				TASK_STATUS,
				COUNT(TASK_STATUS) AS STATUS_COUNT,
				(
					COUNT(TASK_STATUS)::FLOAT / (
						SELECT
							COUNT(*)
						FROM
							LOGS
					)
				) * 100 AS PERCENTAGE
			FROM
				LOGS
			GROUP BY
				TASK_STATUS;
		`

		fmt.Println("[query]: ", q)
		rows, err := db.Query(q)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't worng while fetching status summary."})
			return
		}

		defer rows.Close()
		type Summary struct {
			TaskStatus  string  `json:"taskStatus"`
			StatusCount int     `json:"statusCount"`
			Percentage  float64 `json:"percentage"`
		}

		var summary []Summary
		for rows.Next() {
			var row Summary
			err := rows.Scan(
				&row.TaskStatus,
				&row.StatusCount,
				&row.Percentage,
			)

			if err != nil {
				http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
				return
			}

			summary = append(summary, row)
		}

		if summary == nil {
			json.NewEncoder(w).Encode(struct {
				Summary []Summary `json:"statusSummary"`
			}{Summary: []Summary{}})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Summary []Summary `json:"statusSummary"`
		}{Summary: summary})
	})

	mux.HandleFunc("GET /type-summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := `
			SELECT
				TASK_TYPE,
				COUNT(TASK_TYPE) AS TYPE_COUNT,
				(
					COUNT(TASK_TYPE)::FLOAT / (
						SELECT
							COUNT(*)
						FROM
							LOGS
					)
				) * 100 AS PERCENTAGE
			FROM
				LOGS
			GROUP BY
				TASK_TYPE;
		`

		fmt.Println("[query]: ", q)
		rows, err := db.Query(q)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't worng while fetching task type summary."})
			return
		}

		defer rows.Close()
		type Summary struct {
			TaskType    string  `json:"taskType"`
			StatusCount int     `json:"statusCount"`
			Percentage  float64 `json:"percentage"`
		}

		var summary []Summary
		for rows.Next() {
			var row Summary
			err := rows.Scan(
				&row.TaskType,
				&row.StatusCount,
				&row.Percentage,
			)

			if err != nil {
				http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
				return
			}

			summary = append(summary, row)
		}

		if summary == nil {
			json.NewEncoder(w).Encode(struct {
				Summary []Summary `json:"typeSummary"`
			}{Summary: []Summary{}})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Summary []Summary `json:"typeSummary"`
		}{Summary: summary})
	})

	mux.HandleFunc("GET /daily-task-count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := `
			SELECT
				CREATED_AT::DATE AS CREATED_DATE,
				TO_CHAR(CREATED_AT::DATE, 'DD MON YYYY') AS FORMATTED_DATE,
				COUNT(*) AS TASK_COUNT
			FROM
				LOGS
			GROUP BY
				CREATED_DATE
			ORDER BY
				CREATED_DATE;
		`
		// Print the query
		fmt.Println("[query]: ", q)

		// Execute the query
		rows, err := db.Query(q)

		// Error handling
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't wrong while fetching daily task count"})
			return
		}

		// Close the rows
		defer rows.Close()

		// define the response structure
		type DailyTask struct {
			CreatedDate   time.Time `json:"createdDate"`
			FormattedDate string    `json:"formattedDate"`
			TaskCount     int       `json:"taskCount"`
		}

		dailyTasks := []DailyTask{}

		// Scan the rows
		for rows.Next() {
			var row DailyTask
			err := rows.Scan(
				&row.CreatedDate,
				&row.FormattedDate,
				&row.TaskCount,
			)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": "Something wen't wrong while fetching daily task count"})
				return
			}

			dailyTasks = append(dailyTasks, row)
		}

		if len(dailyTasks) == 0 {
			json.NewEncoder(w).Encode(struct {
				DailyTasks []DailyTask `json:"dailyTasks"`
			}{DailyTasks: []DailyTask{}})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			DailyTasks []DailyTask `json:"dailyTasks"`
		}{DailyTasks: dailyTasks})
	})

	mux.HandleFunc("GET /completed-task-count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		view := r.URL.Query().Get("v")
		duration := r.URL.Query().Get("d")

		if view == "" {
			view = "week"
		}

		if duration == "" {
			duration = "1 months"
		}

		var interval string
		if view == "week" {
			interval = "1 week"
		} else {
			interval = "1 month"
		}

		q := fmt.Sprintf(`
			WITH
				DATE_SERIES AS (
					SELECT
						GENERATE_SERIES(
							DATE_TRUNC('%s', NOW() - INTERVAL '%s'),
							DATE_TRUNC('%s', NOW()),
							INTERVAL '%s'
						) AS DATE_START
				),
				LOGS_BY_VIEW AS (
					SELECT
						DATE_TRUNC('%s', COMPLETED_AT) AS DATE_START,
						COUNT(TASK_NAME) AS TASK_COUNT
					FROM
						LOGS
					GROUP BY
						DATE_START
				)
			SELECT
				D.DATE_START,
				COALESCE(L.TASK_COUNT, 0) AS TASK_COUNT
			FROM
				DATE_SERIES D
				LEFT JOIN LOGS_BY_VIEW L ON D.DATE_START = L.DATE_START
			ORDER BY
				DATE_START;
		`, view, duration, view, interval, view)

		fmt.Println("[query]: ", q)

		rows, err := db.Query(q)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "Something wen't wrong while getting completed task count", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		type CompletedCount struct {
			CompletedAt time.Time `json:"completedAt"`
			TaskCount   int       `json:"taskCount"`
		}

		var completedCounts []CompletedCount
		for rows.Next() {
			var completedCount CompletedCount
			rows.Scan(
				&completedCount.CompletedAt,
				&completedCount.TaskCount,
			)

			completedCounts = append(completedCounts, completedCount)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Message        string           `json:"message"`
			CompletedCount []CompletedCount `json:"completedCount"`
		}{Message: "ok", CompletedCount: completedCounts})
	})

	mux.HandleFunc("GET /task-summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := `
			SELECT
				COUNT(*) AS TOTAL_TASKS,
				(
					SELECT
						COUNT(*)
					FROM
						LOGS
					WHERE
						TASK_TYPE = 'bug'
				) AS TOTAL_BUGS,
				(
					SELECT
						COUNT(*)
					FROM
						LOGS
					WHERE
						TASK_STATUS = 'progress'
				) AS TOTAL_PROGRESS_TASKS,
				(
					SELECT
						COUNT(*)
					FROM
						LOGS
					WHERE
						PRIORITY = 10
				) AS HIGHEST_PRIORITY_TASKS
			FROM
				LOGS;
		`

		rows, err := db.Query(q)
		if err != nil {
			http.Error(w, "Something wen't wrong while generate summary", http.StatusInternalServerError)
			return
		}

		defer rows.Close()
		for rows.Next() {

		}

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":"+serverPort, corsMiddleware(mux)); err != nil {
		panic(err)
	}
}
