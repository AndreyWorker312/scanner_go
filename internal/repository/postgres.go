package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"network-scanner/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(cfg struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}) (*PostgresRepository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) SaveScanRequest(ctx context.Context, req *models.ScanRequest) (int64, error) {
	query := `INSERT INTO scan_requests (ip_address, ports) VALUES ($1, $2) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, req.IPAddress, req.Ports).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to save scan request: %w", err)
	}
	return id, nil
}

func (r *PostgresRepository) SaveScanResults(ctx context.Context, results []*models.ScanResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO scan_results (request_id, port, is_open) VALUES ($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, result := range results {
		if _, err = stmt.ExecContext(ctx, result.RequestID, result.Port, result.IsOpen); err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetScanHistory(ctx context.Context) ([]*models.ScanResponse, error) {
	query := `
		SELECT 
			r.id, 
			r.ip_address, 
			r.ports, 
			r.created_at, 
			COALESCE(array_agg(sr.port) FILTER (WHERE sr.is_open = true), '{}'::int[]) as open_ports
		FROM scan_requests r
		LEFT JOIN scan_results sr ON r.id = sr.request_id
		GROUP BY r.id
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan history: %w", err)
	}
	defer rows.Close()

	var history []*models.ScanResponse
	for rows.Next() {
		var resp models.ScanResponse
		var req models.ScanRequest
		var openPorts []int64

		if err := rows.Scan(
			&req.ID,
			&req.IPAddress,
			&req.Ports,
			&req.CreatedAt,
			pq.Array(&openPorts),
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		resp.Request = &req
		resp.OpenPorts = make([]int, len(openPorts))
		for i, port := range openPorts {
			resp.OpenPorts[i] = int(port)
		}

		history = append(history, &resp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return history, nil
}

func (r *PostgresRepository) GetScanResults(ctx context.Context, requestID int64) (*models.ScanResponse, error) {
	var req models.ScanRequest
	query := `SELECT id, ip_address, ports, created_at FROM scan_requests WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(
		&req.ID,
		&req.IPAddress,
		&req.Ports,
		&req.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get scan request: %w", err)
	}

	resultsQuery := `SELECT port, is_open FROM scan_results WHERE request_id = $1`
	rows, err := r.db.QueryContext(ctx, resultsQuery, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan results: %w", err)
	}
	defer rows.Close()

	var results []*models.ScanResult
	var openPorts []int
	for rows.Next() {
		var res models.ScanResult
		if err := rows.Scan(
			&res.Port,
			&res.IsOpen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}
		res.RequestID = requestID

		results = append(results, &res)
		if res.IsOpen {
			openPorts = append(openPorts, res.Port)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &models.ScanResponse{
		Request:   &req,
		Results:   results,
		OpenPorts: openPorts,
	}, nil
}

func (r *PostgresRepository) Close() error {
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
